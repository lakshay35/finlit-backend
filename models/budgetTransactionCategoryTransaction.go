package models

import "github.com/google/uuid"

// BudgetTransactionCategoryTransaction ...
type BudgetTransactionCategoryTransaction struct {
	TransactionName string `json:"transaction_name"`
	CategoryName    string `json:"category_name"`
}

//BudgetTransactionCategoryTransactionCreationPayload ...
type BudgetTransactionCategoryTransactionCreationPayload struct {
	TransactionName string    `json:"transaction_name"`
	CategoryName    string    `json:"category_name"`
	BudgetID        uuid.UUID `json:"budget_id"`
}
