package models

import "github.com/google/uuid"

// BudgetTransactionCategory ...
type BudgetTransactionCategory struct {
	BudgetTransactionCategoryID uuid.UUID `json:"budget_transaction_category_id"`
	BudgetID                    uuid.UUID `json:"budget_id"`
	CategoryName                string    `json:"category_name"`
}

// BudgetTransactionCategoryCreationPayload ...
type BudgetTransactionCategoryCreationPayload struct {
	BudgetID     uuid.UUID `json:"budget_id"`
	CategoryName string    `json:"category_name"`
}
