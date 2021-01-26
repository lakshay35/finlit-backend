package models

import "github.com/google/uuid"

// AddRolePayload ...
type AddRolePayload struct {
	Role     string    `json:"role"`
	BudgetID uuid.UUID `json:"budgetId"`
}
