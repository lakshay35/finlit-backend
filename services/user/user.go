package user

import (
	"fmt"
	"net/http"

	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	"github.com/lakshay35/finlit-backend/utils/database"
)

//GetUser ...
// Gets user from database
func GetUser(googleID string) (*models.User, *errors.Error) {
	tx := database.GetConnection()
	defer database.CloseConnection(tx)

	stmt := database.PrepareStatement(tx, "SELECT * FROM users where google_id = $1")
	fmt.Println("searching for user with googleID", googleID)
	res, err := stmt.Query(googleID)

	if err != nil || !res.Next() {
		return nil, &errors.Error{
			Message:    "User does not exist",
			StatusCode: http.StatusNotFound,
		}
	}

	var userResult models.User

	err = res.Scan(
		&userResult.UserID,
		&userResult.FirstName,
		&userResult.LastName,
		&userResult.Email,
		&userResult.Phone,
		&userResult.GoogleID,
		&userResult.RegistrationDate,
	)

	if err != nil {
		panic(err)
	}

	return &userResult, nil
}

// RegisterUser ...
// Registers user in the db
func RegisterUser(user models.UserRegistrationPayload) (*models.User, *errors.Error) {
	connection := database.GetConnection()
	defer database.CloseConnection(connection)

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
