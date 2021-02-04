package models

import "github.com/google/uuid"

// BudgetTransactionSource ...
type BudgetTransactionSource struct {
	BudgetTransactionSourceID uuid.UUID `json:"budget_transaction_source_id"`
	ExternalAccountID         uuid.UUID `json:"external_account_id"`
	BudgetID                  uuid.UUID `json:"budget_id"`
}

// BudgetTransactionSourcePayload ...
type BudgetTransactionSourcePayload struct {
	BudgetTransactionSourceID uuid.UUID `json:"budget_transaction_source_id"`
	ExternalAccountID         uuid.UUID `json:"external_account_id"`
	BudgetID                  uuid.UUID `json:"budget_id"`
	AccountName               string    `json:"account_name,omitempty"`
}

// BudgetTransactionSourceCreationPayload ...
type BudgetTransactionSourceCreationPayload struct {
	ExternalAccountID uuid.UUID `json:"external_account_id"`
	BudgetID          uuid.UUID `json:"budget_id"`
}
