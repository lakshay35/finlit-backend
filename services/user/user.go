package user

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	"github.com/lakshay35/finlit-backend/utils/database"
)

type CustomError struct {
	Message string
}

// Error ...
// Return error Message
func (err CustomError) Error() string {
	return err.Message
}

//GetUser ...
// Gets user from database
func GetUser(googleID string) (*models.User, error) {
	tx := database.GetConnection()
	defer tx.Commit()

	stmt := database.PrepareStatement(tx, "SELECT * FROM users where google_id = $1")
	fmt.Println("searching for user with googleID", googleID)
	res, err := stmt.Query(googleID)

	if err != nil || !res.Next() {
		panic(err)
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

// RegisterUser ...
// Registers user in the db
func RegisterUser(user models.UserRegistrationPayload) (*models.User, *errors.Error) {
	connection := database.GetConnection()
	defer connection.Commit()

	stmt := database.PrepareStatement(
		connection,
		"INSERT INTO users (first_name, last_name, email, phone, google_id) VALUES ($1, $2, $3, $4, $5) RETURNING user_id, registration_date",
	)

	var result models.User

	err := stmt.QueryRow(
		result.FirstName,
		result.LastName,
		result.Email,
		result.Phone,
		result.GoogleID,
	).Scan(&result.UserID, &result.RegistrationDate)

	if err != nil {
		return nil, &errors.Error{
			Message:    "User already exists",
			StatusCode: http.StatusConflict,
		}
	}

	return &result, nil
}
