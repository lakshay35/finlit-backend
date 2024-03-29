package models

import "github.com/google/uuid"

// Expense ...
// Represents an expense entity
type Expense struct {
	ExpenseID                    uuid.UUID          `json:"expense_id,omitempty"`
	BudgetID                     uuid.UUID          `json:"budget_id"`
	ExpenseName                  string             `json:"expense_name"`
	ExpenseValue                 float32            `json:"expense_value"`
	ExpenseDescription           string             `json:"expense_description,omitempty"`
	ExpenseChargeCycle           ExpenseChargeCycle `json:"expense_charge_cycle"`
	ExpenseTransactionCategories []string           `json:"expense_transaction_categories"`
}

// AddExpensePayload payload for incoming expense addition requests
type AddExpensePayload struct {
	BudgetID                     uuid.UUID          `json:"budget_id"`
	ExpenseName                  string             `json:"expense_name"`
	ExpenseValue                 float32            `json:"expense_value"`
	ExpenseDescription           string             `json:"expense_description,omitempty"`
	ExpenseChargeCycle           ExpenseChargeCycle `json:"expense_charge_cycle"`
	BudgetTransactionCategoryID  uuid.UUID          `json:"budget_transaction_category_id"`
	ExpenseTransactionCategories []string           `json:"expense_transaction_categories"`
}

// ExpenseBudgetTransactionCategory ...
type ExpenseBudgetTransactionCategory struct {
	ExpenseID                  uuid.UUID `json:"expense_id"`
	BudgeTransactionCategoryID uuid.UUID `json:"budget_transaction_category_id"`
	CategoryName               string    `json:"category_name"`
}
