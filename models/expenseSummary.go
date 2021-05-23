package models

import "github.com/plaid/plaid-go/plaid"

// ExpenseSummary ...
type ExpenseSummary struct {
	ExpenseName            string                   `json:"expense_name"`
	ExpenseChargeCycleDays int                      `json:"expense_charge_cycle_days"`
	ExpenseLimit           float64                  `json:"expense_limit"`
	CurrentExpense         float64                  `json:"current_expense"`
	ExpenseCategories      []ExpenseCategorySummary `json:"categories"`
}

// ExpenseCategorySummary ...
//
type ExpenseCategorySummary struct {
	CategoryName string              `json:"category_name"`
	Transactions []plaid.Transaction `json:"transactions"`
}
