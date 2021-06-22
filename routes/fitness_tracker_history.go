package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/services/fitness_tracker_history"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// GetUserFitnessHistory ...
// @Summary Gets user fitness history
// @Description Get user fitness history records with notes
// @Tags Fitness Tracker
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} []models.FitnessHistoryRecord
// @Failure 403 {object} models.Error
// @Router /fitness-tracker/history [get]
func GetUserFitnessHistory(c *gin.Context) {

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	history := fitness_tracker_history.GetUserFitnessHistory(user.UserID)

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
// @Success 200 {object} []models.FitnessHistoryRecord
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

	checkInError := fitness_tracker_history.CheckIn(user.UserID, payload.ActiveToday, payload.Note)

	if checkInError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			checkInError.Reason,
		)

		return
	}

	c.Status(http.StatusOK)
}
