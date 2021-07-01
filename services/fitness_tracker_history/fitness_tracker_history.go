package fitness_tracker_history

import (
	"time"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/utils/database"
)

// GetUserFitnessHistory...
// Retrieves user fitness history
func GetUserFitnessHistory(userId uuid.UUID, pageIndex int) (*models.FitnessHistory, *models.Error) {
	totalRecords, totalPages := TotalPagesAndRecords(userId)

	if pageIndex >= totalPages || pageIndex < 0 {
		return nil, &models.Error{
			Error:  true,
			Reason: "Page index is out of bounds",
		}
	}
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select active_today, date, note from fitness_tracker_history WHERE user_id = $1 ORDER BY date desc LIMIT 10 OFFSET $2"

	stmt := database.PrepareStatement(conn, query)

	rows, queryError := stmt.Query(userId, pageIndex*10)

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

	return &models.FitnessHistory{
		TotalRecords: totalRecords,
		Records:      result,
		PageIndex:    pageIndex,
		TotalPages:   totalPages,
	}, nil
}

// CheckIn...
// Records user fitness checkin
func CheckIn(userId uuid.UUID, activeToday bool, note string) (*models.FitnessHistoryRecord, *models.Error) {
	if HasUserCheckedIn(userId) {
		return nil, &models.Error{
			Error:  true,
			Reason: "You have already checked in for today",
		}
	}
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "INSERT INTO fitness_tracker_history (active_today, note, user_id, date) VALUES ($1, $2, $3, $4)"

	stmt := database.PrepareStatement(conn, query)

	est, estErr := time.LoadLocation("EST")

	if estErr != nil {
		panic(estErr)
	}

	currentTime := time.Now().In(est).Format("01-02-2006")

	_, execError := stmt.Exec(activeToday, note, userId, currentTime)

	if execError != nil {
		panic(execError)
	}

	return &models.FitnessHistoryRecord{
		ActiveToday: activeToday,
		Date:        time.Now().In(est),
		Note:        note,
	}, nil
}

// HasUserCheckedIn...
// Determines if user has checked in today
func HasUserCheckedIn(userId uuid.UUID) bool {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "SELECT COUNT(*) as count FROM fitness_tracker_history WHERE user_id = $1 AND date = $2"

	stmt := database.PrepareStatement(conn, query)

	var count int

	est, estErr := time.LoadLocation("EST")

	if estErr != nil {
		panic(estErr)
	}

	currentTime := time.Now().In(est).Format("01-02-2006")

	_ = stmt.QueryRow(userId, currentTime).Scan(&count)

	return count > 0
}

// TODO: Index the date field in fitness_tracker_history and define it in db.sql

// TotalCheckinRecords...
// Returns total number of check in records for a given user id
func TotalCheckinRecords(userId uuid.UUID) int {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "SELECT COUNT(*) as count FROM fitness_tracker_history WHERE user_id = $1"

	stmt := database.PrepareStatement(conn, query)

	var count int

	_ = stmt.QueryRow(userId).Scan(&count)

	return count
}

// TotalPages...
// Returns total number of pages a user's fitness history has
func TotalPagesAndRecords(userId uuid.UUID) (int, int) {
	totalRecords := TotalCheckinRecords(userId)

	remainder := totalRecords % 10

	if remainder > 0 {
		return totalRecords, (totalRecords / 10) + 1
	}

	return totalRecords, (totalRecords / 10)
}

func RecentCheckinHistory(userId uuid.UUID) []models.FitnessHistoryRecord {

	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select date, active_today, note from fitness_tracker_history WHERE user_id = $1 ORDER BY date desc LIMIT 5"

	stmt := database.PrepareStatement(conn, query)

	rows, queryErr := stmt.Query(userId)

	if queryErr != nil {
		panic(queryErr)
	}

	recentHistory := make([]models.FitnessHistoryRecord, 0)

	for rows.Next() {
		var record models.FitnessHistoryRecord
		rows.Scan(&record.Date, &record.ActiveToday, &record.Note)
		recentHistory = append(recentHistory, record)
	}

	return recentHistory
}

// GetUserFitnessRate
func GetUserFitnessRate(userId uuid.UUID) float64 {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select active_today from fitness_tracker_history WHERE user_id = $1"

	stmt := database.PrepareStatement(conn, query)

	rows, queryErr := stmt.Query(userId)

	if queryErr != nil {
		panic(queryErr)
	}

	activeCount := 0.0
	total := 0.0

	for rows.Next() {
		var active_today bool

		rows.Scan(&active_today)

		if active_today {
			activeCount++
		}

		total++

	}

	return activeCount / total
}
