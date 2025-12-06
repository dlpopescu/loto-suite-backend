package utils

import (
	"fmt"
	"loto-suite/backend/generics"
	"loto-suite/backend/models"
	"strconv"
)

func CheckBilet649(checkResult *models.CheckResult) {
	game, _ := GetGameById("649")
	VerificareNoroc649(checkResult.LuckyNumber, checkResult.DrawResult.LuckyNumber, game.LuckyNumberDigitCount, game.LuckyNumberMinMatchLen)
	varianteJucateLen := len(checkResult.Numbers)

	if varianteJucateLen == 0 {
		variantaJucata := models.Variant{
			Numbers:     []models.Number{},
			WinsRegular: getDefaultCategoriiCastigVariante649(),
			WinsSpecial: getDefaultCategoriiCastigVariante649(),
		}

		checkResult.Numbers = append(checkResult.Numbers, variantaJucata)
	} else {
		for i := range checkResult.Numbers {
			// Reset Castigator flags once before verification
			for j := range checkResult.Numbers[i].Numbers {
				checkResult.Numbers[i].Numbers[j].IsWinner = false
			}

			VerificareVarianta649(&checkResult.Numbers[i], checkResult.DrawResult.VariantRegular, game.VariantMinNumbersCount, game.VariantDrawNumbersCount)
			VerificareVarianta649(&checkResult.Numbers[i], checkResult.DrawResult.VariantSpecial, game.VariantMinNumbersCount, game.VariantDrawNumbersCount)
		}
	}
}

func VerificareVarianta649(variantaJucata *models.Variant, variantaExtrasa *models.Variant, minNumerePerVariantaJucata int, numerePerVariantaExtrasa int) {
	// Don't reset Castigator flags - they might have been set by previous verification calls

	isValidTicket := len(variantaJucata.Numbers) >= minNumerePerVariantaJucata
	isValidDraw := variantaExtrasa.Id != -1 && len(variantaExtrasa.Numbers) == numerePerVariantaExtrasa

	if !isValidTicket || !isValidDraw {
		switch variantaExtrasa.Id {
		case 1:
			variantaJucata.WinsRegular = getDefaultCategoriiCastigVariante649()
		case 2:
			variantaJucata.WinsSpecial = getDefaultCategoriiCastigVariante649()
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

	matchingCount := 0
	for i := range variantaJucata.Numbers {
		for _, numarCastigator := range variantaExtrasa.Numbers {
			if variantaJucata.Numbers[i].Value == numarCastigator.Value {
				variantaJucata.Numbers[i].IsWinner = true
				matchingCount++
				break
			}
		}
	}

	for cnt := 6; cnt >= 3; cnt-- {
		castig := models.Win{
			Id:          fmt.Sprintf("%v (%v/%v)", generics.RomanNumbers[6-cnt+1], cnt, numerePerVariantaExtrasa),
			Description: fmt.Sprintf("%v numere din cele %v extrase", cnt, numerePerVariantaExtrasa),
		}

		if cnt == 6 {
			castig.IsWinner = matchingCount >= cnt
		} else {
			castig.IsWinner = matchingCount == cnt
		}

		switch variantaExtrasa.Id {
		case 1:
			variantaJucata.WinsRegular = append(variantaJucata.WinsRegular, castig)
		case 2:
			variantaJucata.WinsSpecial = append(variantaJucata.WinsSpecial, castig)
		}
	}
}

func VerificareNoroc649(norocJucat *models.LuckyNumber, norocCastigator *models.LuckyNumber, norocLen int, minNorocMatchLen int) {
	foundWinner := false

	numarNorocJucat := norocJucat.Value
	numarNorocCastigator := norocCastigator.Value

	for n := norocLen; n >= minNorocMatchLen; n-- {
		descriere := ""
		if n == norocLen {
			descriere = fmt.Sprintf("Toate cele %v cifre ale numarului (in ordine)", n)
		} else {
			descriere = fmt.Sprintf("Ultimele %v cifre ale numarului (in ordine)", n)
		}

		isWinner := false

		if !foundWinner {
			if len(numarNorocJucat) == len(numarNorocCastigator) && len(numarNorocJucat) == norocLen {
				isWinner = numarNorocJucat[norocLen-n:] == numarNorocCastigator[norocLen-n:]
			}
		}

		norocJucat.Wins = append(norocJucat.Wins,
			models.Win{
				Id:          fmt.Sprintf("%v", generics.RomanNumbers[norocLen-n+1]),
				Description: descriere,
				IsWinner:    !foundWinner && isWinner,
			})

		if isWinner {
			foundWinner = true
		}
	}

	castigatorNPlus3 := false
	castigatorNMinus3 := false

	numarNorocJucatInt, err1 := strconv.Atoi(numarNorocJucat)
	numarNorocCastigatorInt, err2 := strconv.Atoi(numarNorocCastigator)
	if (err1 == nil) && (err2 == nil) {
		castigatorNPlus3 = numarNorocJucatInt == numarNorocCastigatorInt+3
		castigatorNMinus3 = numarNorocJucatInt == numarNorocCastigatorInt-3
	}

	norocJucat.Wins = append(norocJucat.Wins, models.Win{
		Id:          "N+3",
		Description: "Tot numarul + 3",
		IsWinner:    castigatorNPlus3,
	})

	norocJucat.Wins = append(norocJucat.Wins, models.Win{
		Id:          "N-3",
		Description: "Tot numarul - 3",
		IsWinner:    castigatorNMinus3,
	})

	norocJucat.IsWinner = foundWinner || castigatorNPlus3 || castigatorNMinus3
}

func getDefaultCategoriiCastigVariante649() []models.Win {
	castiguri := []models.Win{}

	for cnt := 6; cnt >= 3; cnt-- {
		castiguri = append(castiguri, models.Win{
			Id:          fmt.Sprintf("%v (%v/%v)", generics.RomanNumbers[6-cnt+1], cnt, 6),
			Description: fmt.Sprintf("%v numere din cele %v extrase", cnt, 6),
			IsWinner:    false,
		})
	}

	return castiguri
}
