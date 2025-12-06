package models

type Win struct {
	Id          string `json:"id"`
	Description string `json:"descriere"`
	IsWinner    bool   `json:"castigator"`
}

type WinCumulated struct {
	Id          string  `json:"id"`
	Description string  `json:"descriere"`
	WinCount    int     `json:"win_count"`
	Value       float64 `json:"valoare"`
}

type WinCategory struct {
	Id    string  `json:"id_categorie"`
	Value float64 `json:"valoare_castig"`
}
