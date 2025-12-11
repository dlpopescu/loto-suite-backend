package models

type LuckyNumber struct {
	Value    string `json:"numar"`
	IsWinner bool   `json:"castigator,omitempty"`
	Wins     []Win  `json:"-"`
	// Wins     []Win  `json:"castiguri,omitempty"`
}

type Variant struct {
	Id          int      `json:"id"`
	Numbers     []Number `json:"numere,omitempty"`
	WinsRegular []Win    `json:"-"`
	WinsSpecial []Win    `json:"-"`
	// WinsRegular []Win    `json:"castiguri,omitempty"`
	// WinsSpecial []Win    `json:"castiguri_varianta_speciala,omitempty"`
}

type Number struct {
	Value    int  `json:"numar"`
	IsWinner bool `json:"castigator,omitempty"`
}
