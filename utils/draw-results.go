package utils

import (
	"encoding/json"
	"fmt"
	"loto-suite/backend/cache"
	"loto-suite/backend/models"
	"time"
)

func GetDrawResults(gameId string, month string, year string) ([]models.DrawResult, error) {
	if gameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	if cachedData, found := cache.Get(gameId, month, year); found {
		var results []models.DrawResult
		if err := json.Unmarshal(cachedData, &results); err == nil {
			return results, nil
		}
	}

	game, err := GetGameById(gameId)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	results, err := scrapeDrawResults(game, month, year)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if data, marshalErr := json.Marshal(results); marshalErr == nil {
		cache.Set(gameId, month, year, data, 24*time.Hour)
	}

	return results, nil
}
