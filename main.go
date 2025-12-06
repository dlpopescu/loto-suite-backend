package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"loto-suite/backend/generics"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"loto-suite/backend/utils"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	mux *http.ServeMux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	server := NewServer()

	addr := fmt.Sprintf(":%s", port)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ips, err := getLocalIPs()
		if err != nil {
			logging.InfoBe(fmt.Sprintf("Server ready at http://localhost%s (could not get local IPs)", addr))
		} else {
			for _, ip := range ips {
				logging.InfoBe(fmt.Sprintf("Server ready at http://%s%s", ip, addr))
			}
			// logging.InfoBe("If you want to access this server from outside your network, set up port forwarding on your router for this port, and allow it through your firewall.")
		}
	}()
	// ---
	// To make this server available from anywhere:
	// 1. Ensure your router forwards external traffic on the chosen port (e.g. 8888) to this machine's local IP.
	// 2. Allow incoming connections on this port in your firewall (System Preferences > Security > Firewall).
	// 3. Access from outside: http://<your-public-ip>:8888
	// 4. For production, use HTTPS and a reverse proxy (Nginx, Caddy, etc).
	// ---

	if err := http.ListenAndServe(addr, server.mux); err != nil {
		log.Fatal(err)
	}
}

func getLocalIPs() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}

	s.mux.HandleFunc("/api/games", corsMiddleware(s.handleGetGames))
	s.mux.HandleFunc("/api/draw-dates", corsMiddleware(s.handleGetDrawDates))
	s.mux.HandleFunc("/api/draw-results", corsMiddleware(s.handleGetDrawResults))
	s.mux.HandleFunc("/api/check", corsMiddleware(s.handleVerificareBilet))
	s.mux.HandleFunc("/api/scan", corsMiddleware(s.handleScanareBilet))
	// s.mux.HandleFunc("/api/log", corsMiddleware(s.handleLog))
	// s.mux.HandleFunc("/api/clear-cache", corsMiddleware(s.handleClearCache))

	return s
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleGetGames(w http.ResponseWriter, r *http.Request) {
	games := models.Games
	respondWithJSON(w, games)
}

func (s *Server) handleGetDrawDates(w http.ResponseWriter, r *http.Request) {
	daysBackStr := r.URL.Query().Get("days_back")
	daysBack := 60
	if daysBackStr != "" {
		if days, err := strconv.Atoi(daysBackStr); err == nil {
			daysBack = days
		}
	}

	dates := utils.GetDrawDates(daysBack)
	respondWithJSON(w, dates)
}

func (s *Server) handleGetDrawResults(w http.ResponseWriter, r *http.Request) {
	queryGameId := strings.TrimSpace(r.URL.Query().Get("game"))
	queryDateStr := strings.TrimSpace(r.URL.Query().Get("date"))

	if queryGameId == "" || queryDateStr == "" {
		respondWithError(w, "missing game or date parameter", http.StatusBadRequest, "fe")
		return
	}

	queryDate, err := generics.TryParseDate(queryDateStr)
	if err != nil {
		respondWithError(w, "invalid date format", http.StatusBadRequest, "fe")
		return
	}

	month := strconv.Itoa(int(queryDate.Month()))
	year := strconv.Itoa(queryDate.Year())

	var body struct {
		UseCache bool `json:"use_cache,omitempty"`
	}

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil && err != io.EOF {
		respondWithError(w, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	drawResults, err := utils.GetDrawResults(queryGameId, month, year, body.UseCache)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	result, _ := generics.FindFirst(drawResults, func(dr models.DrawResult) bool {
		date, err := generics.TryParseDate(dr.GameDate)
		return err == nil && date.Equal(queryDate)
	})

	if result.GameId != "" {
		respondWithJSON(w, result)
		return
	}

	respondWithError(w, "nu am gasit rezultate pentru data selectata", http.StatusNotFound, "fe")
}

func (s *Server) handleVerificareBilet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "method not allowed", http.StatusMethodNotAllowed, "fe")
		return
	}

	req := models.CheckRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	result, err := utils.CheckTicket(req)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	respondWithJSON(w, result)
}

func (s *Server) handleScanareBilet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "method not allowed", http.StatusMethodNotAllowed, "fe")
		return
	}

	var req struct {
		GameId    string `json:"game_id"`
		ImageData string `json:"image_data"` // Base64 encoded
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		respondWithError(w, "invalid base64 image data", http.StatusBadRequest, "fe")
		return
	}

	result, err := utils.ScanareBilet(req.GameId, imageData)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	respondWithJSON(w, result)
}

// func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		logging.ErrorFe("na", fmt.Sprintf("method not allowed %s", r.Method))
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var logReq struct {
// 		Level      string `json:"level"`
// 		Message    string `json:"message"`
// 		CallerInfo string `json:"caller_info"`
// 		StackTrace string `json:"stack_trace"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil {
// 		respondWithError(w, fmt.Sprintf("invalid body: %s", err.Error()), http.StatusBadRequest, "fe")
// 		return
// 	}

// 	// Route to appropriate logging function based on level
// 	level := strings.ToLower(logReq.Level)
// 	switch level {
// 	case "debug":
// 		logging.DebugFe(logReq.CallerInfo, logReq.Message)
// 	case "info":
// 		logging.InfoFe(logReq.CallerInfo, logReq.Message)
// 	case "warn":
// 		logging.WarnFe(logReq.CallerInfo, logReq.Message)
// 	case "error":
// 		logging.ErrorFe(logReq.CallerInfo, logReq.Message)
// 	case "fatal":
// 		logging.FatalFe(logReq.CallerInfo, logReq.Message, logReq.StackTrace)
// 	default:
// 		logging.InfoFe(logReq.CallerInfo, logReq.Message)
// 	}

// 	w.WriteHeader(http.StatusOK)
// }

// func (s *Server) handleClearCache(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	cache.ClearCache()

// 	w.WriteHeader(http.StatusOK)
// }

func respondWithJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, message string, status int, source string) {
	switch strings.ToLower(source) {
	case "fe":
		logging.ErrorFe(message, "")
	default:
		logging.ErrorBe(message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
