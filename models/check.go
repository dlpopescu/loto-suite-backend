package models

type CheckRequest struct {
	GameId      string    `json:"game_id"`
	Variants    []Variant `json:"variante"`
	LuckyNumber string    `json:"noroc,omitempty"`
	Date        string    `json:"date"`
	UseCache    bool      `json:"use_cache"`
}

type CheckResult struct {
	DrawResult                  *DrawResult    `json:"draw_result"`
	Numbers                     []Variant      `json:"variante_jucate"`
	LuckyNumber                 *LuckyNumber   `json:"noroc_jucat,omitempty"`
	WinsCumulatedVariantRegular []WinCumulated `json:"castiguri_varianta"`
	WinsCumulatedVariantSpecial []WinCumulated `json:"castiguri_varianta_speciala"`
	WinsCumulatedLuckyNumber    []WinCumulated `json:"castiguri_noroc"`
	WinsTotal                   float64        `json:"valoare_totala_castig"`
}
