package utils

import (
	"fmt"
	"loto-suite/backend/generics"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"strconv"
	"strings"
)

func CheckTicket(request models.CheckRequest) (*models.CheckResult, error) {
	if len(request.Variants) == 0 {
		return nil, fmt.Errorf("at least one set of numbers is required")
	}

	requestDate, dateParseErr := generics.TryParseDate(request.Date)
	if dateParseErr != nil {
		return nil, fmt.Errorf("invalid date format")
	}

	month := strconv.Itoa(int(requestDate.Month()))
	year := strconv.Itoa(requestDate.Year())

	logging.InfoBe(fmt.Sprintf("Checking %s", generics.SerializeIgnoreError(request)))

	request.GameId = strings.ToLower(strings.TrimSpace(request.GameId))
	drawResults, err := GetDrawResults(request.GameId, month, year, request.UseCache)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to get draw results: %v", err)
	}

	drawResult, _ := generics.FindFirst(drawResults, func(dr models.DrawResult) bool {
		date, err := generics.TryParseDate(dr.GameDate)
		return err == nil && date.Equal(requestDate)
	})

	if drawResult.GameId == "" {
		return nil, fmt.Errorf("no draw results found for the specified date")
	}

	checkResult := models.CheckResult{
		DrawResult: &drawResult,
		Numbers:    request.Variants,
	}

	request.LuckyNumber = strings.TrimSpace(request.LuckyNumber)
	if request.LuckyNumber != "" {
		checkResult.LuckyNumber = &models.LuckyNumber{
			Value: request.LuckyNumber,
		}
	}

	switch request.GameId {
	case "649":
		CheckBilet649(&checkResult)
	case "540":
		CheckBilet540(&checkResult)
	case "joker":
		CheckBiletJoker(&checkResult)
	}

	checkResult.WinsTotal = 0
	checkResult.WinsCumulatedVariantRegular = []models.WinCumulated{}
	checkResult.WinsCumulatedVariantSpecial = []models.WinCumulated{}

	for _, varianta := range checkResult.Numbers {
		for _, castig := range varianta.WinsRegular {
			castigIndex := generics.IndexOf(
				checkResult.WinsCumulatedVariantRegular,
				func(c models.WinCumulated) bool {
					return c.Id == castig.Id
				})

			if castigIndex == -1 {
				castigCumulat := models.WinCumulated{
					Id:          castig.Id,
					Description: castig.Description,
					WinCount:    generics.Btoi(castig.IsWinner),
					Value:       0,
				}

				valoareCastig, found := generics.FindFirst(
					drawResult.WinCategoriesVariantRegular,
					func(v models.WinCategory) bool {
						return v.Id == castig.Id
					})

				if found {
					castigCumulat.Value = valoareCastig.Value
				}

				checkResult.WinsCumulatedVariantRegular = append(checkResult.WinsCumulatedVariantRegular, castigCumulat)
			} else {
				checkResult.WinsCumulatedVariantRegular[castigIndex].WinCount += generics.Btoi(castig.IsWinner)
			}
		}

		for _, castig := range varianta.WinsSpecial {
			castigIndex := generics.IndexOf(
				checkResult.WinsCumulatedVariantSpecial,
				func(c models.WinCumulated) bool {
					return c.Id == castig.Id
				})

			if castigIndex == -1 {
				castigCumulat := models.WinCumulated{
					Id:          castig.Id,
					Description: castig.Description,
					WinCount:    generics.Btoi(castig.IsWinner),
					Value:       0,
				}

				valoareCastig, found := generics.FindFirst(
					drawResult.WinCategoriesVariantSpecial,
					func(v models.WinCategory) bool {
						return v.Id == castig.Id
					})

				if found {
					castigCumulat.Value = valoareCastig.Value
				}

				checkResult.WinsCumulatedVariantSpecial = append(checkResult.WinsCumulatedVariantSpecial, castigCumulat)
			} else {
				checkResult.WinsCumulatedVariantSpecial[castigIndex].WinCount += generics.Btoi(castig.IsWinner)
			}
		}
	}

	for _, castigVarianta := range checkResult.WinsCumulatedVariantRegular {
		checkResult.WinsTotal += float64(castigVarianta.WinCount) * castigVarianta.Value
	}

	for _, castigVarianta := range checkResult.WinsCumulatedVariantSpecial {
		checkResult.WinsTotal += float64(castigVarianta.WinCount) * castigVarianta.Value
	}

	checkResult.WinsCumulatedLuckyNumber = []models.WinCumulated{}
	for _, castig := range checkResult.LuckyNumber.Wins {
		castigIndex := generics.IndexOf(
			checkResult.WinsCumulatedLuckyNumber,
			func(c models.WinCumulated) bool {
				return c.Id == castig.Id
			})

		if castigIndex == -1 {
			castigCumulat := models.WinCumulated{
				Id:          castig.Id,
				Description: castig.Description,
				WinCount:    generics.Btoi(castig.IsWinner),
				Value:       0,
			}

			valoareCastig, found := generics.FindFirst(
				drawResult.WinCategoriesLuckyNumber,
				func(v models.WinCategory) bool {
					return v.Id == castig.Id
				})

			if found {
				castigCumulat.Value = valoareCastig.Value
			}

			checkResult.WinsCumulatedLuckyNumber = append(checkResult.WinsCumulatedLuckyNumber, castigCumulat)
		} else {
			checkResult.WinsCumulatedLuckyNumber[castigIndex].WinCount += generics.Btoi(castig.IsWinner)
		}
	}

	for _, castigNoroc := range checkResult.WinsCumulatedLuckyNumber {
		checkResult.WinsTotal += float64(castigNoroc.WinCount) * castigNoroc.Value
	}

	// v, _ := json.Marshal(checkResult)
	// fmt.Println(string(v))

	return &checkResult, nil
}
