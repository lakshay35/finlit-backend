package models

import "github.com/google/uuid"

// Budget ...
type Budget struct {
	BudgetName string    `json:"budget_name,omitepty"`
	BudgetID   uuid.UUID `json:"budget_id,omitempty"`
	OwnerID    uuid.UUID `json:"owner_id,omitempty"`
}

// CreateBudgetPayload ...
type CreateBudgetPayload struct {
	BudgetName string `json:"budget_name"`
}
