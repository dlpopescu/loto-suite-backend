package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"loto-suite/backend/generics"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func doHttpRequest(ctx context.Context, method string, url string, customHeaders map[string]string, body url.Values) (*http.Response, error) {
	bodyEncoded := body.Encode()

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(bodyEncoded))
	if err != nil {
		err := fmt.Errorf("failed to %s: %v", method, err)
		logging.ErrorBe(err.Error())

		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0.1 Safari/605.1.15")

	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	logData := map[string]any{
		"method":  method,
		"url":     url,
		"headers": req.Header,
		"body":    bodyEncoded,
	}

	logJson, err := json.Marshal(logData)
	if err == nil {
		logging.DebugBe(fmt.Sprintf("HTTP Request: %s", string(logJson)))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err := fmt.Errorf("failed to %s: %v", method, err)
		logging.ErrorBe(err.Error())

		return nil, err
	}

	return resp, nil
}

func ContainsNumarByValue(slice []models.Number, item models.Number) bool {
	for _, element := range slice {
		if element.Value == item.Value {
			return true
		}
	}

	return false
}

func GetDrawDates(daysBack int) []models.DrawDate {
	now := time.Now()
	loc := now.Location()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	drawTime := time.Date(now.Year(), now.Month(), now.Day(), 21, 00, 0, 0, loc)

	maxBackDate := today.AddDate(0, 0, -daysBack)

	dates := make([]models.DrawDate, 0, 32)

	for d := today; !d.Before(maxBackDate); d = d.AddDate(0, 0, -1) {
		wd := d.Weekday()
		if wd != time.Thursday && wd != time.Sunday {
			continue
		}

		if d.Equal(today) && now.Before(drawTime) {
			continue
		}

		value := d.Format(generics.GoDateFormat)
		zi, ok := generics.DrawDays[int(wd)]
		if !ok {
			zi = generics.DayNames[int(wd)]
		}

		dates = append(dates, models.DrawDate{
			Value:        value,
			DisplayLabel: fmt.Sprintf("%s - %s", d.Format(generics.DateDisplayFormat), zi),
		})
	}

	return dates
}

func GetGameById(gameId string) (*models.Game, error) {
	for _, game := range models.Games {
		if game.Id == gameId {
			return game, nil
		}
	}

	return nil, fmt.Errorf("unsupported game: %s (use 649, 540, or joker)", gameId)
}
