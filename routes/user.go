package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	userService "github.com/lakshay35/finlit-backend/services/user"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// RegisterUser ...
// @Summary Registers user to the database
// @Description Registers a user profile in the finlit database
// @Tags Users
// @Accept  json
// @Produce  json
// @Param body body models.UserRegistrationPayload true "User Information Paylod"
// @Security Google AccessToken
// @Success 200 {object} models.User
// @Failure 403 {object} models.Error
// @Failure 409 {object} models.Error
// @Router /user/register [post]
func RegisterUser(c *gin.Context) {

	var json models.UserRegistrationPayload
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	user, userRegistrationError := userService.RegisterUser(json)

	if userRegistrationError != nil {
		requests.ThrowError(
			c,
			userRegistrationError.StatusCode,
			userRegistrationError.Error(),
		)
	}

	c.JSON(http.StatusOK, user)
}

// GetUserProfile ...
// @Summary Gets user from the database
// @Description Gets the user's profile from the finlit database
// @Tags Users
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.User
// @Failure 403 {object} models.Error
// @Failure 404 {object} models.Error
// @Router /user/get [get]
func GetUserProfile(c *gin.Context) {
	user := requests.GetUserFromContext(c)

	res, err := userService.GetUser(user.GoogleID)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusNotFound,
			"Unable to find user profile registration",
		)

		return
	}

	c.JSON(http.StatusOK, res)
}
