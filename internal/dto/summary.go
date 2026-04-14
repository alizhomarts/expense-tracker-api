package dto

type SummaryResponse struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}

type CategorySummaryResponse struct {
	CategoryID   string  `json:"category_id,omitempty"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
}

type MonthlySummaryResponse struct {
	Month        string  `json:"month"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}
