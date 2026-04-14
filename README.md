# 🚀 Expense Tracker API

![Go](https://img.shields.io/badge/Go-1.26-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-DB-blue)
![JWT](https://img.shields.io/badge/Auth-JWT-green)
![AI](https://img.shields.io/badge/AI-OpenAI-purple)

REST API for personal finance management built with **Go, Echo, PostgreSQL, and OpenAI**.

---

## 📌 Overview

Expense Tracker API is a backend service for managing personal finances.

It supports:
- transactions
- categories
- budgets
- summaries
- 🤖 AI-powered parsing, insights, and receipt processing

---

## ✨ Features

### 🔐 Authentication
- User registration & login
- JWT access & refresh tokens
- Protected routes

### 💸 Transactions
- Create / read / update / delete transactions
- User-specific data isolation

### 🗂 Categories
- Create and list categories
- Auto-create categories via AI

### 📊 Summary
- Total income / expenses / balance
- Monthly summary
- Category breakdown

### 💰 Budgets
- Monthly budgets per category
- Budget status tracking
- Remaining amount & usage %
- Exceeded budget detection

---

## 🤖 AI Features

### 🧠 Transaction parsing

Convert text into structured transaction:

```json
"Yandex Taxi 1800"

↓

{
  "intent": "expense",
  "amount": 1800,
  "category": "transport",
  "description": "Yandex Taxi ride"
}
🧾 Receipt parsing
Magnum
Хлеб 350
Молоко 620
Курица 2800
Итого 3770

↓

{
  "merchant": "Magnum",
  "total_amount": 3770,
  "items": [...]
}
⚡ Receipt → Transaction

Automatically creates transaction from receipt:

POST /api/v1/ai/receipt-to-transaction
📈 Insights
Spending analysis
AI-generated financial insights
🚨 Budget alerts
AI-generated warnings
Detect overspending
Suggest improvements
🧱 Tech Stack
Go
Echo
PostgreSQL
pgx
JWT
Logrus
OpenAI API
📂 Project Structure
cmd/app              - entry point
db/migrations        - database migrations
internal/

  apperror/          - custom errors
  config/            - configuration
  db/                - DB connection
  dto/               - request/response models
  entity/            - domain entities
  http/              - handlers, router, middleware
  logger/            - logging
  repository/        - DB layer
  service/           - business logic
🔗 API Endpoints
Auth
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
User
GET /api/v1/users/me
Transactions
POST /api/v1/transactions
GET /api/v1/transactions
GET /api/v1/transactions/:id
PUT /api/v1/transactions/:id
DELETE /api/v1/transactions/:id
Categories
POST /api/v1/categories
GET /api/v1/categories
Summary
GET /api/v1/summary
GET /api/v1/summary/categories
GET /api/v1/summary/monthly?month=YYYY-MM
Budgets
POST /api/v1/budgets
GET /api/v1/budgets
GET /api/v1/budgets/status?month=YYYY-MM
AI
POST /api/v1/ai/parse
POST /api/v1/ai/parse-and-create
GET /api/v1/ai/insights?month=YYYY-MM
GET /api/v1/ai/budget-alerts?month=YYYY-MM
POST /api/v1/ai/parse-receipt
POST /api/v1/ai/receipt-to-transaction
⚙️ Environment

Create .env:

APP_PORT=8989

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=expense_tracker

JWT_ACCESS_SECRET=your_access_secret
JWT_REFRESH_SECRET=your_refresh_secret

OPENAI_API_KEY=your_api_key
OPENAI_MODEL=gpt-5.3-chat-latest
▶️ Run
go run ./cmd/app
🧪 Future Improvements
📦 Receipt → multiple transactions
🖼 OCR for receipt images
📄 Swagger documentation
🐳 Docker support
✅ Unit tests
📊 Reports export
🤖 AI finance assistant

## 👨‍💻 Author

**Alizhomart Shukayev**
GitHub: https://github.com/alizhomarts