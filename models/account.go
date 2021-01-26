package models

import "github.com/google/uuid"

// Account ...
// Bank Account Integration Entity
type Account struct {
	ExternalAccountID uuid.UUID `json:"external_account_id"`
	AccountName       string    `json:"account_name,omitempty"`
	AccessToken       string    `json:"access_token,omitempty"`
}
