package models

import "github.com/plaid/plaid-go/plaid"

// ExpenseSummary ...
type ExpenseSummary struct {
	ExpenseCategories []ExpenseCategorySummary `json:"categories"`
}

// ExpenseCategorySummary ...
//
type ExpenseCategorySummary struct {
	CategoryName   string              `json:"category_name"`
	ExpenseLimit   float64             `json:"expense_limit"`
	CurrentExpense float64             `json:"current_expense"`
	Transactions   []plaid.Transaction `json:"transactions"`
}
