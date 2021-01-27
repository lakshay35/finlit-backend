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

	query := "SELECT * FROM budgets WHERE owner = $1 AND budget_id = $2"

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer rows.Close()
	defer database.CloseConnection(connection)

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
	defer database.CloseConnection(connection)

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

// IsUserAdmin ...
// Determines if user is an admin on the given budget
func IsUserAdmin(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	query := `SELECT * FROM user_roles ur WHERE ur.user_id = $1 AND ur.role_id = (SELECT role_id FROM roles WHERE role_name = 'Full Rights')
	AND ur.budget_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		panic(err)
	}

	defer rows.Close()
	defer database.CloseConnection(connection)

	return rows.Next()
}

// IsUserViewer ...
// Determines if user is an admin on the given budget
func IsUserViewer(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	query := `SELECT * FROM user_roles ur WHERE ur.user_id = $1 AND ur.role_id = (SELECT role_id FROM roles WHERE role_name = 'View Rights')
	AND ur.budget_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		panic(err)
	}

	defer rows.Close()
	defer database.CloseConnection(connection)

	return rows.Next()
}

// IsUserOwner ...
// Checks if user is an owner
func IsUserOwner(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	query := `SELECT owner_id FROM budgets WHERE budget_id = $1 AND owner_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(budgetID, userID)

	if err != nil {
		panic(err)
	}

	defer rows.Close()
	defer database.CloseConnection(connection)

	return rows.Next()
}
