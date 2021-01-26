package models

import "github.com/google/uuid"

// Expense ...
// Represents an expense entity
type Expense struct {
	ExpenseID          uuid.UUID          `json:"expense_id,omitempty"`
	BudgetID           uuid.UUID          `json:"budget_id"`
	ExpenseName        string             `json:"expense_name"`
	ExpenseValue       float32            `json:"expense_value"`
	ExpenseDescription string             `json:"expense_description,omitempty"`
	ExpenseChargeCycle ExpenseChargeCycle `json:"expense_charge_cycle"`
}
