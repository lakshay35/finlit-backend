package user

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
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
