package models

import "github.com/google/uuid"

// User ...
type User struct {
	UserID           uuid.UUID `json:"userID,omitempty"`
	RegistrationDate string    `json:"registrationDate,omitempty"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	GoogleID         string    `json:"googleID"`
}

type UserRegistrationPayload struct {
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	GoogleID         string    `json:"googleID"`
}