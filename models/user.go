package models

import "github.com/google/uuid"

// User ...
type User struct {
	UserID           uuid.UUID `json:"user_id,omitempty"`
	RegistrationDate string    `json:"registration_date,omitempty"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	GoogleID         string    `json:"google_id"`
}

// UserRegistrationPayload ...
type UserRegistrationPayload struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	GoogleID  string `json:"google_id"`
}
