package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"loto-suite/backend/generics"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"loto-suite/backend/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	mux *http.ServeMux
}

type contextKey string

const traceIDKey contextKey = "traceID"

func main() {
	srv := NewServer()

	port := os.Getenv("HTTPS_PORT")
	log.Printf("Starting HTTPS server on %s...", port)
	err := http.ListenAndServeTLS(
		fmt.Sprintf(":%s", port),
		os.Getenv("TLS_CERT_FILE"),
		os.Getenv("TLS_KEY_FILE"),
		srv.mux,
	)

	if err != nil {
		log.Fatal(err)
	}
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
	s.mux.HandleFunc("/api/logs", corsMiddleware(s.handleDownloadLogs))
	// s.mux.HandleFunc("/api/log", corsMiddleware(s.handleLog))
	// s.mux.HandleFunc("/api/clear-cache", corsMiddleware(s.handleClearCache))
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

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

		traceMiddleware(next)(w, r)
	}
}

func traceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		w.Header().Set("X-Trace-ID", traceID)

		logging.Info("be", fmt.Sprintf("[TraceID: %s] %s request to %s", traceID, r.Method, r.URL.String()))

		next(w, r.WithContext(ctx))
	}
}

func (s *Server) handleGetGames(w http.ResponseWriter, r *http.Request) {
	games := models.Games
	respondWithJSON(w, r, games)
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
	respondWithJSON(w, r, dates)
}

func (s *Server) handleGetDrawResults(w http.ResponseWriter, r *http.Request) {
	queryGameId := strings.TrimSpace(r.URL.Query().Get("game"))
	queryDateStr := strings.TrimSpace(r.URL.Query().Get("date"))

	if queryGameId == "" || queryDateStr == "" {
		respondWithError(w, r, "missing game or date parameter", http.StatusBadRequest, "fe")
		return
	}

	queryDate, err := generics.TryParseDate(queryDateStr)
	if err != nil {
		respondWithError(w, r, "invalid date format", http.StatusBadRequest, "fe")
		return
	}

	month := strconv.Itoa(int(queryDate.Month()))
	year := strconv.Itoa(queryDate.Year())

	var body struct {
		UseCache bool `json:"use_cache,omitempty"`
	}

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil && err != io.EOF {
		respondWithError(w, r, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	drawResults, err := utils.GetDrawResults(queryGameId, month, year, body.UseCache)
	if err != nil {
		respondWithError(w, r, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	result, _ := generics.FindFirst(drawResults, func(dr models.DrawResult) bool {
		date, err := generics.TryParseDate(dr.GameDate)
		return err == nil && date.Equal(queryDate)
	})

	if result.GameId != "" {
		respondWithJSON(w, r, result)
		return
	}

	respondWithError(w, r, "nu am gasit rezultate pentru data selectata", http.StatusNotFound, "fe")
}

func (s *Server) handleVerificareBilet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, r, "method not allowed", http.StatusMethodNotAllowed, "fe")
		return
	}

	req := models.CheckRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, r, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	// logging.Info("BE", fmt.Sprintf("%s payload: %s", r.Method, generics.SerializeIgnoreError(req)))

	result, err := utils.CheckTicket(req)
	if err != nil {
		respondWithError(w, r, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	respondWithJSON(w, r, result)
}

func (s *Server) handleScanareBilet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, r, "method not allowed", http.StatusMethodNotAllowed, "fe")
		return
	}

	var req struct {
		GameId    string `json:"game_id"`
		ImageData string `json:"image_data"` // Base64 encoded
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, r, "invalid request body", http.StatusBadRequest, "fe")
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		respondWithError(w, r, "invalid base64 image data", http.StatusBadRequest, "fe")
		return
	}

	result, err := utils.ScanareBilet(req.GameId, imageData)
	if err != nil {
		respondWithError(w, r, err.Error(), http.StatusInternalServerError, "be")
		return
	}

	respondWithJSON(w, r, result)
}

func (s *Server) handleDownloadLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, r, "method not allowed", http.StatusMethodNotAllowed, "be")
		return
	}

	queryDateStr := strings.TrimSpace(r.URL.Query().Get("date"))
	if queryDateStr == "" {
		queryDateStr = time.Now().Format(generics.GoDateFormat)
	}

	queryDate, err := generics.TryParseDate(queryDateStr)
	if err != nil {
		respondWithError(w, r, fmt.Sprintf("invalid date format: %s", queryDateStr), http.StatusBadRequest, "fe")
		return
	}

	fileName := fmt.Sprintf("be_%s.log", queryDate.Format(generics.GoDateFormat))
	filePath := filepath.Join(logging.GetLogDir(), fileName)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			respondWithError(w, r, "log file not found", http.StatusNotFound, "be")
		} else {
			respondWithError(w, r, "failed to open log file", http.StatusInternalServerError, "be")
		}

		return
	}

	defer file.Close()

	traceID, _ := r.Context().Value(traceIDKey).(string)
	logging.Info("be", fmt.Sprintf("[TraceID: %s] Success response", traceID))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("failed to stream log file %s: %v", fileName, err)
	}
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, data any) {
	traceID, _ := r.Context().Value(traceIDKey).(string)
	logging.Info("be", fmt.Sprintf("[TraceID: %s] Success response", traceID))

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	response := map[string]any{
		"data":     data,
		"trace_id": traceID,
	}

	encoder.Encode(response)
}

func respondWithError(w http.ResponseWriter, r *http.Request, message string, status int, source string) {
	traceID, _ := r.Context().Value(traceIDKey).(string)
	logMsg := fmt.Sprintf("[TraceID: %s] %s", traceID, message)

	logging.Error(source, errors.New(logMsg), "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":    message,
		"trace_id": traceID,
	})
}
