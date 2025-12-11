package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"loto-suite/backend/cache"
	"loto-suite/backend/generics"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func scrapeDrawResults(game *models.Game, month string, year string) ([]models.DrawResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	form := url.Values{}
	form.Set("select-year", year)
	form.Set("select-month", month)

	resp, err := doHttpRequest(ctx, "POST", game.Url, nil, form)
	if err != nil {
		logging.Error("scrape", err, "")
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Error("scrape", err, "")
		fmt.Println(err)
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var drawResults []models.DrawResult

	divVarianteNormale := doc.Find(".rezultate-extrageri-content.resultDiv").Not(".floatright").Not(".resultspecialDiv")
	divVarianteNormale.Each(func(i int, div *goquery.Selection) {
		scrapedDateStr := strings.TrimSpace(div.Find(".button-open-details span").Text())
		scrapedDate, err := time.Parse(generics.ScrapeDateFormat, scrapedDateStr)

		if err != nil {
			logging.Error("scrape", err, "")
			return
		}

		gameResult := models.DrawResult{
			GameId:                      game.Id,
			GameDate:                    scrapedDate.Format(generics.GoDateFormat),
			WinCategoriesVariantRegular: extractCategoriiCastigVariante(div),
			WinCategoriesLuckyNumber:    extractCategoriiCastigNoroc(div.Next(), game.Id),
			LuckyNumberName:             game.LuckyNumberName,
			LuckyNumber: &models.LuckyNumber{
				Value: strings.TrimSpace(div.Next().Find(".numere-extrase-noroc span").Text()),
			},
			VariantRegular: &models.Variant{
				Id:      1,
				Numbers: extractNumereVarianta(div),
			},
		}

		divVariantaSpeciala := div.PrevFiltered(".rezultate-extrageri-content.resultspecialDiv")
		if divVariantaSpeciala.Length() > 0 {
			gameResult.WinCategoriesVariantSpecial = extractCategoriiCastigVariante(divVariantaSpeciala)
			gameResult.VariantSpecial = &models.Variant{
				Id:      2,
				Numbers: extractNumereVarianta(divVariantaSpeciala),
			}
		}

		drawResults = append(drawResults, gameResult)
	})

	if data, err := json.Marshal(drawResults); err == nil {
		cache.Set(game.Id, month, year, data, 30*time.Minute)
	}

	return drawResults, err
}

func extractNumereVarianta(div *goquery.Selection) []models.Number {
	numbers := []models.Number{}

	div.Find(".numere-extrase img").Each(func(j int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if !exists {
			return
		}

		scrapedNumber := strings.TrimSuffix(path.Base(src), ".png")
		if numar, err := strconv.Atoi(scrapedNumber); err == nil {
			numbers = append(numbers, models.Number{
				Value: numar,
			})
		}
	})

	return numbers
}

func extractCategoriiCastigVariante(div *goquery.Selection) []models.WinCategory {
	categoriiCastig := []models.WinCategory{}

	table := div.Find(".results-table")
	if table.Length() == 0 {
		logging.Info("scrape", "No .results-table found in div")
		return categoriiCastig
	}

	var valoareCastigHeaderColumnIndex int = -1
	table.Find("thead tr th").Each(func(i int, th *goquery.Selection) {
		headerText := strings.TrimSpace(th.Text())
		if strings.Contains(strings.ToLower(headerText), "valoare castig") {
			valoareCastigHeaderColumnIndex = i
			return
		}
	})

	if valoareCastigHeaderColumnIndex == -1 {
		valoareCastigHeaderColumnIndex = 2
	}

	table.Find("tbody tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() > valoareCastigHeaderColumnIndex+1 {
			valoareStr := strings.TrimSpace(tds.Eq(valoareCastigHeaderColumnIndex).Text())
			if valoareStr == "-" || valoareStr == "" {
				valoareStr = "0"
			}

			valoare, err := strToEnglishFloat(valoareStr)
			if err == nil {
				categoriiCastig = append(categoriiCastig, models.WinCategory{
					Id:     strings.TrimSpace(tds.Eq(0).Text()),
					Amount: valoare,
				})
			}
		}
	})

	return categoriiCastig
}

func extractCategoriiCastigNoroc(div *goquery.Selection, gameId string) []models.WinCategory {
	categoriiCastig := []models.WinCategory{}

	table := div.Find(".results-table")
	if table.Length() == 0 {
		logging.Info("scrape", "No .results-table found in div")
		return categoriiCastig
	}

	var valoareCastigHeaderColumnIndex int = -1
	table.Find("thead tr th").Each(func(i int, th *goquery.Selection) {
		headerText := strings.TrimSpace(th.Text())
		if strings.Contains(strings.ToLower(headerText), "valoare castig") {
			valoareCastigHeaderColumnIndex = i
			return
		}
	})

	table.Find("tbody tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")

		valoareCastigTdIndex := valoareCastigHeaderColumnIndex
		reportTdIndex := valoareCastigHeaderColumnIndex + 1

		if gameId != "649" {
			if _, hasColspan := tds.Eq(1).Attr("colspan"); !hasColspan {
				valoareCastigTdIndex = valoareCastigHeaderColumnIndex + 1
				reportTdIndex = valoareCastigHeaderColumnIndex + 2
			}
		}

		if tds.Length() <= valoareCastigTdIndex || tds.Length() <= reportTdIndex {
			return
		}

		valoareStr := strings.TrimSpace(tds.Eq(valoareCastigTdIndex).Text())
		if valoareStr == "-" || valoareStr == "" {
			valoareStr = "0"
		}

		valoare, err := strToEnglishFloat(valoareStr)
		if err == nil {
			categoriiCastig = append(categoriiCastig, models.WinCategory{
				Id:     strings.TrimSpace(tds.Eq(0).Text()),
				Amount: valoare,
			})
		}
	})

	return categoriiCastig
}

func strToEnglishFloat(valoareStr string) (float64, error) {
	englishFloatFormat := strings.ReplaceAll(valoareStr, ".", "")
	englishFloatFormat = strings.ReplaceAll(englishFloatFormat, ",", ".")

	valoare, err := strconv.ParseFloat(englishFloatFormat, 64)
	return valoare, err
}
