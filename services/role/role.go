package role

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
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
