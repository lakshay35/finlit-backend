package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type User struct {
	UserID           uuid.UUID `json:"userID,omitempty"`
	RegistrationDate string    `json:"registrationDate,omitempty`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	GoogleID         string    `json:"googleID"`
}

// RegisterUser ...
// registers user in the database
func RegisterUser(c *gin.Context) {

	res := User{}
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(jsonData, &res)
	tx := GetConnection()
	defer tx.Commit()
	stmt := PrepareStatement(tx, "INSERT INTO users (first_name, last_name, email, phone, google_id) VALUES ($1, $2, $3, $4, $5)")

	_, err = stmt.Exec(res.FirstName, res.LastName, res.Email, res.Phone, res.GoogleID)
	if err != nil {
		ThrowError(c, http.StatusConflict, "User already exists")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": res})
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
func GetUser(googleID string) (*User, error) {
	tx := GetConnection()
	defer tx.Commit()
	fmt.Println("Querying database for " + googleID)

	stmt := PrepareStatement(tx, "SELECT * FROM users where google_id = $1")

	res, err := stmt.Query(googleID)

	if err != nil || !res.Next() {
		fmt.Println("USER NOT FOUND IN DB")
		return nil, CustomError{
			Message: "User does not exist",
		}
	}

	var userResult User

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
