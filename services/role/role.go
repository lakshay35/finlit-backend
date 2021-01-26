package role

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models/errors"
	"github.com/lakshay35/finlit-backend/utils/database"
)

// DoesUserOwnBudget ...
// Determine if given userID is
// the owner of the given budgetID
func DoesUserOwnBudget(userID uuid.UUID, budgetID uuid.UUID) bool {
	connection := database.GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner = $1 AND budget_id = $2"

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return rows.Next()
}

// GetRole ...
// Returns role string
func GetRole(role string) string {
	switch strings.ToUpper(role) {
	case "FULL RIGHTS":
		return "Full Rights"
	default:
		return "View Rights"
	}
}

// AddRoleToBudget ...
// Adds role to budget
func AddRoleToBudget(userID uuid.UUID, budgetID uuid.UUID, role string) *errors.Error {
	// Only budget owner can add users to budget
	if !DoesUserOwnBudget(userID, budgetID) {
		return &errors.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "Not enough permissions to add users",
		}
	}

	connection := database.GetConnection()
	defer connection.Commit()

	query := "INSERT INTO user_roles (user_id, role_id, budget_id) VALUES ('$1, (select role_id from roles where role_name = $2), $3)"

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(userID, GetRole(role), budgetID)

	if err != nil {
		return &errors.Error{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
	}

	return nil
}
