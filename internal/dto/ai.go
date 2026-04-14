package dto

type AIParseRequest struct {
	Text string `json:"text"`
}

type AIParseResponse struct {
	Intent      string  `json:"intent"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}
