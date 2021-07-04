package models

import (
	"time"
)

type FitnessHistoryRecord struct {
	Date        time.Time `json:"date"`
	ActiveToday bool      `json:"active_today"`
	Note        string    `json:"note"`
	FutureDate  bool      `json:"future_date,omitempty"`
}
