package utils

import (
	"fmt"
	"loto-suite/backend/generics"
	"loto-suite/backend/models"
)

func CheckBilet540(checkResult *models.CheckResult) {
	game, _ := GetGameById("540")
	VerificareNoroc540(checkResult.LuckyNumber, checkResult.DrawResult.LuckyNumber, game.LuckyNumberDigitCount, game.LuckyNumberMinMatchLen)
	varianteJucateLen := len(checkResult.Numbers)

	if varianteJucateLen == 0 {
		variantaJucata := models.Variant{
			Numbers:     []models.Number{},
			WinsRegular: getDefaultCategoriiCastigVariante540(),
			WinsSpecial: getDefaultCategoriiCastigVariante540(),
		}

		checkResult.Numbers = append(checkResult.Numbers, variantaJucata)
	} else {
		for i := range checkResult.Numbers {
			for j := range checkResult.Numbers[i].Numbers {
				checkResult.Numbers[i].Numbers[j].IsWinner = false
			}

			VerificareVarianta540(&checkResult.Numbers[i], checkResult.DrawResult.VariantRegular, game.VariantMinNumbersCount, game.VariantDrawNumbersCount)
			VerificareVarianta540(&checkResult.Numbers[i], checkResult.DrawResult.VariantSpecial, game.VariantMinNumbersCount, game.VariantDrawNumbersCount)
		}
	}
}

func VerificareVarianta540(variantaJucata *models.Variant, variantaExtrasa *models.Variant, minNumerePerVariantaJucata int, numerePerVariantaExtrasa int) {
	isValidTicket := len(variantaJucata.Numbers) >= minNumerePerVariantaJucata
	isValidDraw := variantaExtrasa.Id != -1 && len(variantaExtrasa.Numbers) == numerePerVariantaExtrasa

	if !isValidTicket || !isValidDraw {
		switch variantaExtrasa.Id {
		case 1:
			variantaJucata.WinsRegular = getDefaultCategoriiCastigVariante540()
		case 2:
			variantaJucata.WinsSpecial = getDefaultCategoriiCastigVariante540()
		}

		return
	}

	for i := range variantaJucata.Numbers {
		for _, numarExtras := range variantaExtrasa.Numbers {
			if variantaJucata.Numbers[i].Value == numarExtras.Value {
				variantaJucata.Numbers[i].IsWinner = true
				break
			}
		}
	}

	firstFiveWinningNumbers := variantaExtrasa.Numbers[:5]
	matchCount := 0
	for _, n := range firstFiveWinningNumbers {
		matchCount += generics.Btoi(ContainsNumarByValue(variantaJucata.Numbers, n))
	}

	winnerFound := matchCount == 5
	castig := models.Win{
		Id:          fmt.Sprintf("%v (%v/%v*)", generics.RomanNumbers[1], 5, numerePerVariantaExtrasa),
		Description: fmt.Sprintf("%v numere din primele %v numere extrase", 5, 5),
		IsWinner:    winnerFound,
	}

	switch variantaExtrasa.Id {
	case 1:
		variantaJucata.WinsRegular = append(variantaJucata.WinsRegular, castig)
	case 2:
		variantaJucata.WinsSpecial = append(variantaJucata.WinsSpecial, castig)
	}

	matchCount = 0
	for _, n := range variantaExtrasa.Numbers {
		matchCount += generics.Btoi(ContainsNumarByValue(variantaJucata.Numbers, n))
	}

	winnerFound = !winnerFound && matchCount == 5
	castig = models.Win{
		Id:          fmt.Sprintf("%v (%v/%v)", generics.RomanNumbers[2], 5, numerePerVariantaExtrasa),
		Description: fmt.Sprintf("%v numere din toate cele %v numere extrase", 5, numerePerVariantaExtrasa),
		IsWinner:    winnerFound,
	}

	switch variantaExtrasa.Id {
	case 1:
		variantaJucata.WinsRegular = append(variantaJucata.WinsRegular, castig)
	case 2:
		variantaJucata.WinsSpecial = append(variantaJucata.WinsSpecial, castig)
	}

	winnerFound = !winnerFound && matchCount == 4
	castig = models.Win{
		Id:          fmt.Sprintf("%v (%v/%v)", generics.RomanNumbers[3], 5-1, numerePerVariantaExtrasa),
		Description: fmt.Sprintf("%v numere din toate cele %v numere extrase", 5-1, numerePerVariantaExtrasa),
		IsWinner:    winnerFound,
	}

	switch variantaExtrasa.Id {
	case 1:
		variantaJucata.WinsRegular = append(variantaJucata.WinsRegular, castig)
	case 2:
		variantaJucata.WinsSpecial = append(variantaJucata.WinsSpecial, castig)
	}
}

func VerificareNoroc540(norocJucat *models.LuckyNumber, norocCastigator *models.LuckyNumber, norocLen int, minNorocMatchLen int) {
	castiguri := []models.Win{}

	foundWinner := false

	for n := norocLen; n >= minNorocMatchLen; n-- {
		descriere := ""
		if n == norocLen {
			descriere = fmt.Sprintf("Toate cele %v cifre ale numarului (in ordine)", n)
		} else {
			descriere = fmt.Sprintf("Primele sau ultimele %v cifre ale numarului (in ordine)", n)
		}

		isWinner := false

		if !foundWinner {
			numarNorocCastigator := norocCastigator.Value
			if len(norocJucat.Value) == len(numarNorocCastigator) && len(norocJucat.Value) == norocLen {
				isWinner = norocJucat.Value[:n] == numarNorocCastigator[:n] || norocJucat.Value[norocLen-n:] == numarNorocCastigator[norocLen-n:]
			}
		}

		castiguri = append(castiguri,
			models.Win{
				Id:          fmt.Sprintf("%v", generics.RomanNumbers[norocLen-n+1]),
				Description: descriere,
				IsWinner:    !foundWinner && isWinner,
			})

		if isWinner {
			foundWinner = true
		}
	}

	norocJucat.Wins = castiguri
	// Mark the entire Noroc as winner if any category won
	norocJucat.IsWinner = foundWinner
}

func getDefaultCategoriiCastigVariante540() []models.Win {
	castiguri := []models.Win{}
	castiguri = append(castiguri,
		models.Win{
			Id:          "I (5/6*)",
			Description: "5 numere din primele 5 extrase",
			IsWinner:    false,
		})

	castiguri = append(castiguri,
		models.Win{
			Id:          "II (5/6)",
			Description: "5 numere din toate cele 6 numere extrase",
			IsWinner:    false,
		})

	castiguri = append(castiguri,
		models.Win{
			Id:          "III (4/6)",
			Description: "4 numere din toate cele 6 numere extrase",
			IsWinner:    false,
		})

	return castiguri
}
