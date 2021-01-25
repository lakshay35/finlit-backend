package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	userService "github.com/lakshay35/finlit-backend/services/user"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// RegisterUser ...
// @Summary Registers user to the database
// @Description Registers a user profile in the finlit database
// @Tags Users
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.User
// @Failure 403 {object} models.Error
// @Failure 409 {object} models.Error
// @Router /user/register [post]
func RegisterUser(c *gin.Context) {

	res := models.User{}
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(jsonData, &res)
	tx := database.GetConnection()
	defer tx.Commit()
	stmt := database.PrepareStatement(tx, "INSERT INTO users (first_name, last_name, email, phone, google_id) VALUES ($1, $2, $3, $4, $5) RETURNING user_id, registration_date")

	err = stmt.QueryRow(res.FirstName, res.LastName, res.Email, res.Phone, res.GoogleID).Scan(&res.UserID, &res.RegistrationDate)
	if err != nil {
		fmt.Println(err)
		requests.ThrowError(c, http.StatusConflict, "User already exists")
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetUserProfile ...
// @Summary Gets user from the database
// @Description Gets the user's profile from the finlit database
// @Tags Users
// @ID user-get
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.User
// @Failure 403 {object} models.Error
// @Failure 404 {object} models.Error
// @Router /user/get [get]
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("USERID")

	if !exists {
		panic("USERID not present when trying to get user profile")
	}

	user, err := userService.GetUser(userID.(string))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Unable to find user profile registration",
			"error":   true,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CustomError ...
type CustomError struct {
	Message string
}

// Error ...
// Return error Message
func (err CustomError) Error() string {
	return err.Message
}
