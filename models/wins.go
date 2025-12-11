package models

type Win struct {
	Id          string `json:"id"`
	Description string `json:"descriere"`
	IsWinner    bool   `json:"castigator,omitempty"`
}

type WinCumulated struct {
	Id          string  `json:"id"`
	Description string  `json:"descriere"`
	WinCount    int     `json:"win_count"`
	Amount      float64 `json:"suma"`
}

type WinCategory struct {
	Id     string  `json:"id_categorie"`
	Amount float64 `json:"suma"`
}
