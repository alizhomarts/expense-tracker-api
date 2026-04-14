package dto

type AIReceiptParseRequest struct {
	Text string `json:"text"`
}

type ReceiptItemResponse struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

type AIReceiptParseResponse struct {
	Merchant          string                `json:"merchant"`
	TotalAmount       float64               `json:"total_amount"`
	Items             []ReceiptItemResponse `json:"items"`
	SuggestedCategory string                `json:"suggested_category"`
}
