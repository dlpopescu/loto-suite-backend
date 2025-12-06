package models

type Game struct {
	Id                      string `json:"id"`
	DisplayName             string `json:"display_name"`
	Url                     string `json:"url"`
	LuckyNumberDigitCount   int    `json:"numar_cifre_noroc"`
	LuckyNumberMinMatchLen  int    `json:"noroc_min_match_len"`
	VariantMinNumbersCount  int    `json:"min_numere_per_varianta_jucata"`
	VariantsMaxCount        int    `json:"numar_max_variante"`
	VariantDrawNumbersCount int    `json:"numere_per_varianta_extrasa"`
	VariantMinNumber        int    `json:"min_value_numar_varianta"`
	VariantMaxNumber        int    `json:"max_value_numar_varianta"`
	LuckyNumberName         string `json:"nume_noroc"`
}

type LuckyNumber struct {
	Value    string `json:"numar"`
	IsWinner bool   `json:"castigator"`
	Wins     []Win  `json:"castiguri,omitempty"`
}

type Variant struct {
	Id          int      `json:"id"`
	Numbers     []Number `json:"numere,omitempty"`
	WinsRegular []Win    `json:"castiguri,omitempty"`
	WinsSpecial []Win    `json:"castiguri_varianta_speciala,omitempty"`
}

type Number struct {
	Value    int  `json:"numar"`
	IsWinner bool `json:"castigator,omitempty"`
}

type ScanResult struct {
	GameId          string       `json:"game_id"`
	GameDate        string       `json:"game_date"`
	Variants        []Variant    `json:"variante"`
	LuckyNumber     *LuckyNumber `json:"noroc"`
	LuckyNumberName string       `json:"nume_noroc"`
}

type DrawDate struct {
	Value        string `json:"value"`
	DisplayLabel string `json:"label"`
}

type DrawResult struct {
	GameId                      string        `json:"game_id"`
	GameDate                    string        `json:"game_date"`
	VariantRegular              *Variant      `json:"varianta"`
	VariantSpecial              *Variant      `json:"varianta_speciala,omitempty"`
	LuckyNumber                 *LuckyNumber  `json:"noroc"`
	LuckyNumberName             string        `json:"nume_noroc"`
	WinCategoriesVariantRegular []WinCategory `json:"categorii_castig_varianta,omitempty"`
	WinCategoriesVariantSpecial []WinCategory `json:"categorii_castig_varianta_speciala,omitempty"`
	WinCategoriesLuckyNumber    []WinCategory `json:"categorii_castig_noroc,omitempty"`
}
