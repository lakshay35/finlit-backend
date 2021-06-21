package models

type FitnessCheckInPayload struct {
	ActiveToday bool   `json:"active_today"`
	Note        string `json:"note"`
}
