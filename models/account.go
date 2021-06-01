package models

import "github.com/google/uuid"

// Account ...
// Bank Account Integration Entity
type Account struct {
	ExternalAccountID uuid.UUID `json:"external_account_id"`
	AccountName       string    `json:"account_name,omitempty"`
	UserID            uuid.UUID `json:"user_id"`
	AccessToken       string    `json:"access_token,omitempty"`
	InstitutionalID   string    `json:"institutional_id,omitempty"`
}

type AccountIdPayload struct {
	ExternalAccountID uuid.UUID `json:"external_account_id"`
}
