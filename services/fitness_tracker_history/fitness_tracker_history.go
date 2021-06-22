package fitness_tracker_history

import (
	"time"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/logging"
)

func GetUserFitnessHistory(userId uuid.UUID) []models.FitnessHistoryRecord {
	conn := database.GetConnection()

	query := "Select active_today, date, note from fitness_tracker_history WHERE user_id = $1"

	stmt := database.PrepareStatement(conn, query)

	rows, queryError := stmt.Query(userId)

	if queryError != nil {
		panic(queryError)
	}

	result := make([]models.FitnessHistoryRecord, 0)

	for rows.Next() {
		var record models.FitnessHistoryRecord
		scanErr := rows.Scan(&record.ActiveToday, &record.Date, &record.Note)

		if scanErr != nil {
			panic(scanErr)
		}

		result = append(result, record)
	}

	return result
}

func CheckIn(userId uuid.UUID, activeToday bool, note string) *models.Error {
	if HasUserCheckedIn(userId) {
		return &models.Error{
			Error:  true,
			Reason: "You have already checked in for today",
		}
	}
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "INSERT INTO fitness_tracker_history (active_today, note, user_id) VALUES ($1, $2, $3)"

	stmt := database.PrepareStatement(conn, query)

	_, execError := stmt.Exec(activeToday, note, userId)

	if execError != nil {
		panic(execError)
	}

	return nil
}

func HasUserCheckedIn(userId uuid.UUID) bool {
	conn := database.GetConnection()

	query := "SELECT COUNT(*) as count FROM fitness_tracker_history WHERE user_id = $1 AND date = $2"

	stmt := database.PrepareStatement(conn, query)

	var count int

	est, estErr := time.LoadLocation("EST")

	if estErr != nil {
		panic(estErr)
	}

	currentTime := time.Now().In(est).Format("01-02-2006")

	logging.InfoLogger.Println("Using userId " + userId.String() + " and time " + currentTime)

	_ = stmt.QueryRow(userId, currentTime).Scan(&count)

	return count > 0
}
