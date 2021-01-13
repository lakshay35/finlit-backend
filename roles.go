package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type addRoleRequest struct {
	Role     string    `json:"role"`
	BudgetID uuid.UUID `json:"budgetId"`
}

// AddRole ...
// Adds role to user
func AddRole(c *gin.Context) {

	userID, found := c.Get("USERID")
	res := addRoleRequest{}
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		// TODO: Create a request body interceptor annotation
		ThrowError(c, 400, "malformed body struct")
		return
	}

	json.Unmarshal(jsonData, &res)

	if !found {
		panic("User ID not found in request context")
	}

	// Only budget owner can add users to budget
	if !doesUserOwnBudget(userID.(string), res.BudgetID) {
		ThrowError(c, 401, "Not enough permissions to add users to")
	}

	connection := GetConnection()
	defer connection.Commit()

	query := "INSERT INTO user_roles (user_id, role_id, budget_id) VALUES ('$1, (select role_id from roles where role_name = $2), $3)"

	stmt := PrepareStatement(connection, query)

	_, err = stmt.Exec(userID.(string), getRole(res.Role), res.BudgetID)

	if err != nil {
		ThrowError(c, 400, "Unknw")
	}
}

// doesUserOwnBudget
// Determine if given userID is
// the owner of the given budgetID
func doesUserOwnBudget(userID string, budgetID uuid.UUID) bool {
	connection := GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner = $1 AND budget_id = $2"

	stmt := PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return rows.Next()

}

func getRole(role string) string {
	switch strings.ToUpper(role) {
	case "FULL RIGHTS":
		return "Full Rights"
	default:
		return "View Rights"
	}
}
