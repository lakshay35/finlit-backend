package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// RegisterUser ...
// registers user in the database
func RegisterUser(c *gin.Context) {

	res := models.User{}
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(jsonData, &res)
	tx := database.GetConnection()
	defer tx.Commit()
	stmt := database.PrepareStatement(tx, "INSERT INTO users (first_name, last_name, email, phone, google_id) VALUES ($1, $2, $3, $4, $5)")

	_, err = stmt.Exec(res.FirstName, res.LastName, res.Email, res.Phone, res.GoogleID)
	if err != nil {
		fmt.Println(err)
		requests.ThrowError(c, http.StatusConflict, "User already exists")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    res,
		"message": "Successfully registered your profile",
	})
}

// GetUserProfile ...
// Gets user profile from db
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("USERID")

	if !exists {
		panic("USERID not present when trying to get user profile")
	}

	user, err := GetUser(userID.(string))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Unable to find user profile registration",
			"error":   true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully retrieved user profile",
		"data":    user,
	})
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

// GetUser ...
// Gets user object from db
func GetUser(googleID string) (*models.User, error) {
	tx := database.GetConnection()
	defer tx.Commit()
	fmt.Println("Querying database for " + googleID)

	stmt := database.PrepareStatement(tx, "SELECT * FROM users where google_id = $1")

	res, err := stmt.Query(googleID)

	if err != nil || !res.Next() {
		fmt.Println("USER NOT FOUND IN DB")
		return nil, CustomError{
			Message: "User does not exist",
		}
	}

	var userResult models.User

	var user_id uuid.UUID
	var first_name string
	var last_name string
	var email string
	var phone string
	var google_id string
	var registration_date string

	res.Scan(&user_id, &first_name, &last_name, &email, &phone, &google_id, &registration_date)

	userResult.UserID = user_id
	userResult.FirstName = first_name
	userResult.LastName = last_name
	userResult.Email = email
	userResult.Phone = phone
	userResult.GoogleID = google_id
	userResult.RegistrationDate = registration_date

	return &userResult, nil
}
