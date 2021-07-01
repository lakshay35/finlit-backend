package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/services/fitness_tracker_history"
	"github.com/lakshay35/finlit-backend/utils/logging"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// GetUserFitnessHistory ...
// @Summary Gets user fitness history
// @Description Get user fitness history records with notes
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Param page query number true "Page number of record"
// @Security Google AccessToken
// @Success 200 {object} []models.FitnessHistoryRecord
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/history [get]
func GetUserFitnessHistory(c *gin.Context) {
	page := c.Query("page")

	if strings.EqualFold("", page) {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Page number needs to be passed. Note, indexing starts from 0.",
		)

		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	pageIndex, pageIndexParseErr := strconv.Atoi(page)

	if pageIndexParseErr != nil {
		logging.ErrorLogger.Print("Unable to convert " + page + " to an integer using `strconv.Atoi`")
		panic(pageIndexParseErr)
	}

	history, historyErr := fitness_tracker_history.GetUserFitnessHistory(user.UserID, pageIndex)

	if historyErr != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			historyErr.Reason,
		)

		return
	}

	c.JSON(http.StatusOK, history)
}

// CheckIn ...
// @Summary Check in status
// @Description Checks in users status
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Param body body models.FitnessCheckInPayload true "Check-in payload to track user activity status"
// @Security Google AccessToken
// @Success 200 {object} models.FitnessHistoryRecord
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/check-in [post]
func CheckIn(c *gin.Context) {
	var payload models.FitnessCheckInPayload
	err := requests.ParseBody(c, &payload)

	if err != nil {
		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	record, checkInError := fitness_tracker_history.CheckIn(user.UserID, payload.ActiveToday, payload.Note)

	if checkInError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			checkInError.Reason,
		)

		return
	}

	c.JSON(http.StatusOK, record)
}

// CheckInStatus ...
// @Summary Check in status retrieval
// @Description Checks if user has checked in
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {boolean} boolean
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/check-in-status [get]
func CheckInStatus(c *gin.Context) {
	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	hasUserCheckedIn := fitness_tracker_history.HasUserCheckedIn(user.UserID)

	c.JSON(http.StatusOK, hasUserCheckedIn)
}

// GetFitnessRate ...
// @Summary Gets fitness rate for user
// @Description Averages check-ins and gets fitness rate for user
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.FitnessCheckinHistory
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/fitness-rate [get]
func GetFitnessRate(c *gin.Context) {
	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	userFitnessRate := fitness_tracker_history.GetUserFitnessRate(user.UserID)

	c.JSON(http.StatusOK, userFitnessRate)
}

// GetRecentUserFitnessHistory ...
// @Summary Gets user's most recent checkin history
// @Description Retrieves user's most recent 5 checkins
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {array} models.FitnessHistoryRecord
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/recent-history [get]
func GetRecentUserFitnessHistory(c *gin.Context) {
	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	hasUserCheckedIn := fitness_tracker_history.RecentCheckinHistory(user.UserID)

	c.JSON(http.StatusOK, hasUserCheckedIn)
}
