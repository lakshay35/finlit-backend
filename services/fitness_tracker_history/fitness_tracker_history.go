package fitness_tracker_history

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/now"
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

// GetUserCalendarFitnessHistory...
// Retrieves user fitness history
func GetUserCalendarFitnessHistory(userId uuid.UUID, monthIndex int) (*models.FitnessHistory, *models.Error) {

	if monthIndex > 12 || monthIndex < 1 {
		return nil, &models.Error{
			Error:  true,
			Reason: "Month index is out of bounds",
		}
	}
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select active_today, date, note from fitness_tracker_history WHERE user_id = $1 AND date >= $2 AND date <= $3 order by date"

	stmt := database.PrepareStatement(conn, query)

	est, estErr := time.LoadLocation("EST")

	if estErr != nil {
		panic(estErr)
	}

	currentTime := time.Now().In(est)

	date := time.Date(currentTime.Year(), time.Month(monthIndex), 1, 23, 0, 0, 0, est)
	startDate := now.With(date).BeginningOfMonth()
	endDate := now.With(date).EndOfMonth()
	today := time.Now().In(est)

	rows, queryError := stmt.Query(userId, startDate.Format("01-02-2006"), endDate.Format("01-02-2006"))

	if queryError != nil {
		panic(queryError)
	}

	cache := make(map[string]models.FitnessHistoryRecord)

	// Populates map with existing records
	for rows.Next() {
		var record models.FitnessHistoryRecord
		scanErr := rows.Scan(&record.ActiveToday, &record.Date, &record.Note)

		if scanErr != nil {
			panic(scanErr)
		}

		cache[record.Date.Format("01-02-2006")] = record
	}

	result := make([]models.FitnessHistoryRecord, 0)
	currDate := startDate

	// Iterate over days of the month to provide record for each day
	for currDate.Month() == endDate.Month() && currDate.Day() <= endDate.Day() {

		if record, ok := cache[currDate.Format("01-02-2006")]; ok {
			result = append(result, record)
		} else {
			if currDate.Month() > today.Month() {
				result = append(result, models.FitnessHistoryRecord{
					Date:        currDate,
					ActiveToday: false,
					Note:        "Date in Future",
					FutureDate:  true,
				})
			} else if currDate.Month() == today.Month() && currDate.Day() > today.Day() {
				result = append(result, models.FitnessHistoryRecord{
					Date:        currDate,
					ActiveToday: false,
					Note:        "Date in Future",
					FutureDate:  true,
				})
			} else {
				result = append(result, models.FitnessHistoryRecord{
					Date:        currDate,
					ActiveToday: false,
					Note:        "No Check-in Recorded",
					NoCheckin:   true,
				})
			}
		}

		currDate = currDate.AddDate(0, 0, 1)
	}

	return &models.FitnessHistory{
		Month:   monthIndex,
		Records: result,
	}, nil
}

// CheckIn...
// Records user fitness checkin
func CheckIn(userId uuid.UUID, activeToday bool, note string, date *time.Time) (*models.FitnessHistoryRecord, *models.Error) {
	fmt.Println(date)
	if date != nil {
		if HasUserCheckedIn(userId, date) {
			return nil, &models.Error{
				Error:  true,
				Reason: "You have already checked in for " + date.Format("01-02-2006"),
			}
		}
	} else if HasUserCheckedIn(userId, nil) {
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

	selectedDate := time.Now().In(est)

	if date != nil {
		selectedDate = *date
	}

	currentTime := selectedDate.Format("01-02-2006")

	_, execError := stmt.Exec(activeToday, note, userId, currentTime)

	if execError != nil {
		panic(execError)
	}

	return &models.FitnessHistoryRecord{
		ActiveToday: activeToday,
		Date:        selectedDate,
		Note:        note,
	}, nil
}

// HasUserCheckedIn...
// Determines if user has checked in today
func HasUserCheckedIn(userId uuid.UUID, date *time.Time) bool {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "SELECT COUNT(*) as count FROM fitness_tracker_history WHERE user_id = $1 AND date = $2"

	stmt := database.PrepareStatement(conn, query)

	var count int

	if date != nil {
		_ = stmt.QueryRow(userId, date.Format("01-02-2006")).Scan(&count)
	} else {
		est, estErr := time.LoadLocation("EST")

		if estErr != nil {
			panic(estErr)
		}

		currentTime := time.Now().In(est).Format("01-02-2006")
		fmt.Println("checking if user has che3cked in", currentTime)
		_ = stmt.QueryRow(userId, currentTime).Scan(&count)
	}

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
func GetUserFitnessRate(userId uuid.UUID) models.FitnessCheckinHistory {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select active_today from fitness_tracker_history WHERE user_id = $1"

	stmt := database.PrepareStatement(conn, query)

	rows, queryErr := stmt.Query(userId)

	if queryErr != nil {
		panic(queryErr)
	}

	activeCount := 0
	inactiveCount := 0

	for rows.Next() {
		var active_today bool

		rows.Scan(&active_today)

		if active_today {
			activeCount++
		} else {
			inactiveCount++
		}

	}

	return models.FitnessCheckinHistory{
		ActiveCount:   activeCount,
		InactiveCount: inactiveCount,
		TotalCheckins: activeCount + inactiveCount,
	}
}

// GetUserWeeklyFitnessRate
func GetUserWeeklyFitnessRate(userId uuid.UUID) models.FitnessCheckinHistory {
	conn := database.GetConnection()

	defer database.CloseConnection(conn)

	query := "Select active_today from fitness_tracker_history WHERE user_id = $1 ORDER BY date desc LIMIT 7"

	stmt := database.PrepareStatement(conn, query)

	rows, queryErr := stmt.Query(userId)

	if queryErr != nil {
		panic(queryErr)
	}

	activeCount := 0
	inactiveCount := 0

	for rows.Next() {
		var active_today bool

		rows.Scan(&active_today)

		if active_today {
			activeCount++
		} else {
			inactiveCount++
		}

	}

	return models.FitnessCheckinHistory{
		ActiveCount:   activeCount,
		InactiveCount: inactiveCount,
		TotalCheckins: activeCount + inactiveCount,
	}
}
