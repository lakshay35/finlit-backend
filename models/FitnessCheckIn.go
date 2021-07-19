package models

import "time"

type FitnessCheckInPayload struct {
	ActiveToday bool      `json:"active_today"`
	Note        string    `json:"note"`
	Date        time.Time `json:"date,omitempty,string"`
}

type FitnessCheckinHistory struct {
	ActiveCount   int `json:"active_count"`
	InactiveCount int `json:"inactive_count"`
	TotalCheckins int `json:"total_checkins"`
}
