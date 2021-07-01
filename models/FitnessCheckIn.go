package models

type FitnessCheckInPayload struct {
	ActiveToday bool   `json:"active_today"`
	Note        string `json:"note"`
}

type FitnessCheckinHistory struct {
	ActiveCount   int `json:"active_count"`
	InactiveCount int `json:"inactive_count"`
	TotalCheckins int `json:"total_checkins"`
}
