package models

type CheckRequest struct {
	GameId      string    `json:"game_id"`
	LuckyNumber string    `json:"noroc,omitempty"`
	Date        string    `json:"date"`
	Variants    []Variant `json:"variante"`
}

type CheckResult struct {
	IsCastigator                bool           `json:"is_castigator"`
	DrawResult                  *DrawResult    `json:"-"`
	VarianteJucate              []Variant      `json:"variante_jucate"`
	LuckyNumber                 *LuckyNumber   `json:"noroc_jucat,omitempty"`
	WinsCumulatedVariantRegular []WinCumulated `json:"castiguri_varianta,omitempty"`
	WinsCumulatedVariantSpecial []WinCumulated `json:"castiguri_varianta_speciala,omitempty"`
	WinsCumulatedLuckyNumber    []WinCumulated `json:"castiguri_noroc,omitempty"`
	WinsTotal                   float64        `json:"castiguri_total"`
}
