package models

type DrawDate struct {
	Date  string `json:"date"`
	Label string `json:"label"`
}

type DrawResult struct {
	GameId                      string        `json:"game_id"`
	GameDate                    string        `json:"game_date"`
	VariantRegular              *Variant      `json:"varianta"`
	VariantSpecial              *Variant      `json:"varianta_speciala,omitempty"`
	LuckyNumber                 *LuckyNumber  `json:"noroc"`
	LuckyNumberName             string        `json:"nume_noroc"`
	WinCategoriesVariantRegular []WinCategory `json:"-"`
	WinCategoriesVariantSpecial []WinCategory `json:"-"`
	WinCategoriesLuckyNumber    []WinCategory `json:"-"`
	// WinCategoriesVariantRegular []WinCategory `json:"categorii_castig_varianta,omitempty"`
	// WinCategoriesVariantSpecial []WinCategory `json:"categorii_castig_varianta_speciala,omitempty"`
	// WinCategoriesLuckyNumber    []WinCategory `json:"categorii_castig_noroc,omitempty"`
}
