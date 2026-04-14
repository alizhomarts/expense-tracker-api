package service

import (
	"bytes"
	"context"
	"encoding/json"
	"expense-tracker-api/internal/entity"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"

	"github.com/google/uuid"
)

const openAIResponsesURL = "https://api.openai.com/v1/responses"

type AIService struct {
	transactionService *TransactionService
	categoryService    *CategoryService
	summaryService     *SummaryService
	budgetService      *BudgetService
	apiKey             string
	model              string
	httpClient         *http.Client
}

func NewAIService(
	transactionService *TransactionService,
	categoryService *CategoryService,
	summaryService *SummaryService,
	budgetService *BudgetService,
	apiKey string,
	model string,
) *AIService {
	return &AIService{
		transactionService: transactionService,
		categoryService:    categoryService,
		summaryService:     summaryService,
		budgetService:      budgetService,
		apiKey:             apiKey,
		model:              model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *AIService) ParseTransactionText(ctx context.Context, text string) (*dto.AIParseResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, apperror.InvalidRequestBody
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("openai api key is empty")
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"intent": map[string]any{
				"type": "string",
				"enum": []string{"expense", "income"},
			},
			"amount": map[string]any{
				"type": "number",
			},
			"category": map[string]any{
				"type": "string",
			},
			"description": map[string]any{
				"type": "string",
			},
		},
		"required":             []string{"intent", "amount", "category", "description"},
		"additionalProperties": false,
	}

	systemPrompt := `You are a financial transaction parser.
Convert the user's text into valid JSON.

Rules:
- intent must be either "expense" or "income"
- amount must be a positive number
- category must be short and lowercase, for example: food, transport, groceries, salary, shopping, entertainment
- description should be concise and human-readable
- if the text clearly describes spending, use "expense"
- if the text clearly describes received money, use "income"
- return only data that matches the schema`

	requestBody := map[string]any{
		"model": s.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": systemPrompt,
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": text,
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "transaction_parse",
				"strict": true,
				"schema": schema,
			},
		},
	}

	respBytes, err := s.callOpenAI(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	var openAIResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal openai response wrapper: %w", err)
	}

	var resultText string
	for _, out := range openAIResp.Output {
		if out.Type != "message" {
			continue
		}

		for _, content := range out.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				resultText = content.Text
				break
			}
		}

		if resultText != "" {
			break
		}
	}

	if strings.TrimSpace(resultText) == "" {
		return nil, fmt.Errorf("openai response text is empty")
	}

	var parsed dto.AIParseResponse
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal parsed transaction json: %w", err)
	}

	parsed.Intent = strings.TrimSpace(parsed.Intent)
	parsed.Category = strings.TrimSpace(strings.ToLower(parsed.Category))
	parsed.Description = strings.TrimSpace(parsed.Description)

	if parsed.Intent == "" || parsed.Amount <= 0 || parsed.Category == "" || parsed.Description == "" {
		return nil, fmt.Errorf("parsed transaction has invalid fields")
	}

	return &parsed, nil
}

func (s *AIService) ParseAndCreate(ctx context.Context, userID uuid.UUID, text string) (*dto.TransactionResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, apperror.InvalidRequestBody
	}

	parsed, err := s.ParseTransactionText(ctx, text)
	if err != nil {
		return nil, err
	}

	category, err := s.categoryService.GetOrCreate(
		ctx,
		userID,
		parsed.Category,
		entity.TransactionType(parsed.Intent),
	)
	if err != nil {
		return nil, fmt.Errorf("get or create category: %w", err)
	}

	req := &dto.CreateTransactionRequest{
		Type:        parsed.Intent,
		Amount:      parsed.Amount,
		Description: parsed.Description,
		CategoryID:  category.ID,
	}

	return s.transactionService.Create(ctx, userID, req)
}

func (s *AIService) GenerateInsights(ctx context.Context, userID uuid.UUID, month string) (*dto.AIInsightsResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}
	if strings.TrimSpace(month) == "" {
		return nil, apperror.ErrInvalidMonth
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("openai api key is empty")
	}

	monthlySummary, err := s.summaryService.GetMonthlySummary(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get monthly summary for insights: %w", err)
	}

	categorySummary, err := s.summaryService.GetCategorySummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get category summary for insights: %w", err)
	}

	type categoryItem struct {
		CategoryName string  `json:"category_name"`
		Total        float64 `json:"total"`
	}

	categories := make([]categoryItem, 0, len(categorySummary))
	for _, item := range categorySummary {
		categories = append(categories, categoryItem{
			CategoryName: item.CategoryName,
			Total:        item.Total,
		})
	}

	contextPayload := map[string]any{
		"month": month,
		"summary": map[string]any{
			"total_income":  monthlySummary.TotalIncome,
			"total_expense": monthlySummary.TotalExpense,
			"balance":       monthlySummary.Balance,
		},
		"categories": categories,
	}

	contextJSON, err := json.Marshal(contextPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal insights context: %w", err)
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"insights": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required":             []string{"insights"},
		"additionalProperties": false,
	}

	systemPrompt := `You are a personal finance assistant.
Based on the provided monthly finance data, generate 3 to 5 short, useful insights.

Rules:
- Keep insights concise and clear
- Mention the biggest expense categories if possible
- Mention if income is zero
- Mention if balance is negative
- Do not hallucinate missing data
- Return only JSON that matches the schema`

	requestBody := map[string]any{
		"model": s.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": systemPrompt,
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": string(contextJSON),
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "finance_insights",
				"strict": true,
				"schema": schema,
			},
		},
	}

	respBytes, err := s.callOpenAI(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	var openAIResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal openai response wrapper: %w", err)
	}

	var resultText string
	for _, out := range openAIResp.Output {
		if out.Type != "message" {
			continue
		}

		for _, content := range out.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				resultText = content.Text
				break
			}
		}

		if resultText != "" {
			break
		}
	}

	if strings.TrimSpace(resultText) == "" {
		return nil, fmt.Errorf("openai insights response text is empty")
	}

	var parsed dto.AIInsightsResponse
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal insights json: %w", err)
	}

	if len(parsed.Insights) == 0 {
		return nil, fmt.Errorf("insights list is empty")
	}

	return &parsed, nil
}

func (s *AIService) GenerateBudgetAlerts(ctx context.Context, userID uuid.UUID, month string) (*dto.AIBudgetAlertsResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}
	if strings.TrimSpace(month) == "" {
		return nil, apperror.ErrInvalidMonth
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("openai api key is empty")
	}

	budgetStatuses, err := s.budgetService.GetStatus(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get budget status for alerts: %w", err)
	}

	type budgetItem struct {
		CategoryName string  `json:"category_name"`
		BudgetAmount float64 `json:"budget_amount"`
		SpentAmount  float64 `json:"spent_amount"`
		Remaining    float64 `json:"remaining"`
		IsExceeded   bool    `json:"is_exceeded"`
		UsagePercent float64 `json:"usage_percent"`
	}

	items := make([]budgetItem, 0, len(budgetStatuses))
	for _, item := range budgetStatuses {
		items = append(items, budgetItem{
			CategoryName: item.CategoryName,
			BudgetAmount: item.BudgetAmount,
			SpentAmount:  item.SpentAmount,
			Remaining:    item.Remaining,
			IsExceeded:   item.IsExceeded,
			UsagePercent: item.UsagePercent,
		})
	}

	contextPayload := map[string]any{
		"month":   month,
		"budgets": items,
	}

	contextJSON, err := json.Marshal(contextPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal budget alerts context: %w", err)
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"alerts": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required":             []string{"alerts"},
		"additionalProperties": false,
	}

	systemPrompt := `You are a strict personal finance assistant.
Based on the budget data:
- Warn when usage > 80%
- Warn clearly when exceeded
- Suggest action if spending is zero (tracking issue)
- Be short and direct

Examples:
- "Food budget is already at 85%. Slow down spending."
- "Transport budget exceeded by 2,500."
- "No expenses recorded for food. You may not be tracking correctly."

Return 2–5 alerts.`

	requestBody := map[string]any{
		"model": s.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": systemPrompt,
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": string(contextJSON),
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "budget_alerts",
				"strict": true,
				"schema": schema,
			},
		},
	}

	respBytes, err := s.callOpenAI(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	var openAIResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal openai response wrapper: %w", err)
	}

	var resultText string
	for _, out := range openAIResp.Output {
		if out.Type != "message" {
			continue
		}

		for _, content := range out.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				resultText = content.Text
				break
			}
		}

		if resultText != "" {
			break
		}
	}

	if strings.TrimSpace(resultText) == "" {
		return nil, fmt.Errorf("openai budget alerts response text is empty")
	}

	var parsed dto.AIBudgetAlertsResponse
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal budget alerts json: %w", err)
	}

	if len(parsed.Alerts) == 0 {
		return nil, fmt.Errorf("budget alerts list is empty")
	}

	return &parsed, nil
}

func (s *AIService) ParseReceiptText(ctx context.Context, text string) (*dto.AIReceiptParseResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, apperror.InvalidRequestBody
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("openai api key is empty")
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"merchant": map[string]any{
				"type": "string",
			},
			"total_amount": map[string]any{
				"type": "number",
			},
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
						"amount": map[string]any{
							"type": "number",
						},
						"category": map[string]any{
							"type": "string",
						},
					},
					"required":             []string{"name", "amount", "category"},
					"additionalProperties": false,
				},
			},
			"suggested_category": map[string]any{
				"type": "string",
			},
		},
		"required":             []string{"merchant", "total_amount", "items", "suggested_category"},
		"additionalProperties": false,
	}

	systemPrompt := `You are a receipt parser.

Parse receipt text into valid JSON.

Rules:
- merchant: store or merchant name
- total_amount: total receipt amount
- items: list of purchased items
- each item must include name, amount, and short lowercase category
- suggested_category: one main overall category for the whole receipt, for example groceries, food, transport, shopping
- return only data matching the schema
- do not invent extra items
- if item categories are unclear, use "other"`

	requestBody := map[string]any{
		"model": s.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": systemPrompt,
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": text,
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "receipt_parse",
				"strict": true,
				"schema": schema,
			},
		},
	}

	respBytes, err := s.callOpenAI(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	var openAIResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal openai receipt response wrapper: %w", err)
	}

	var resultText string
	for _, out := range openAIResp.Output {
		if out.Type != "message" {
			continue
		}

		for _, content := range out.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				resultText = content.Text
				break
			}
		}

		if resultText != "" {
			break
		}
	}

	if strings.TrimSpace(resultText) == "" {
		return nil, fmt.Errorf("openai receipt response text is empty")
	}

	var parsed dto.AIReceiptParseResponse
	if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal receipt json: %w", err)
	}

	parsed.Merchant = strings.TrimSpace(parsed.Merchant)
	parsed.SuggestedCategory = strings.TrimSpace(strings.ToLower(parsed.SuggestedCategory))

	if parsed.Merchant == "" || parsed.TotalAmount <= 0 {
		return nil, fmt.Errorf("parsed receipt has invalid fields")
	}

	return &parsed, nil
}

func (s *AIService) ReceiptToTransaction(ctx context.Context, userID uuid.UUID, text string) (*dto.TransactionResponse, error) {
	// 1. parse receipt
	parsed, err := s.ParseReceiptText(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("parse receipt: %w", err)
	}

	// 2. category
	category, err := s.categoryService.GetOrCreate(
		ctx,
		userID,
		parsed.SuggestedCategory,
		entity.TransactionTypeExpense,
	)
	if err != nil {
		return nil, fmt.Errorf("get or create category: %w", err)
	}

	// 3. create transaction
	req := &dto.CreateTransactionRequest{
		Type:        string(entity.TransactionTypeExpense),
		Amount:      parsed.TotalAmount,
		Description: fmt.Sprintf("%s receipt", parsed.Merchant),
		CategoryID:  category.ID,
	}

	return s.transactionService.Create(ctx, userID, req)
}

func (s *AIService) callOpenAI(ctx context.Context, requestBody map[string]any) ([]byte, error) {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIResponsesURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send openai request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}
