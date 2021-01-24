package budget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// ParseBudget ...
// Parses body to budget{} type
// Throws error if body does not match
func ParseBudget(c *gin.Context, res *models.Budget) error {
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonData, &res)

	if err != nil {
		requests.ThrowError(c, 400, "request body structure match error")
		return err
	}

	return nil
}

// DoesBudgetExist ...
// Checks if a budget exists
func DoesBudgetExist(UserID uuid.UUID, budgetName string) bool {
	connection := database.GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner_id = $1 AND budget_name = $2"

	stmt, err := connection.Prepare(query)

	if err != nil {
		return false
	}

	res, err := stmt.Query(UserID, budgetName)

	if err != nil {
		panic(err.Error())
	}
	return res.Next()
}

// GetBudget ...
// Gets budget from db based
// on given params
func GetBudget(userID uuid.UUID, budgetName string) models.Budget {
	connection := database.GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner_id = $1 AND budget_name = $2"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Something went wrong when preparing query to get budget")
	}

	rows, err := stmt.Query(userID, budgetName)

	if err != nil || !rows.Next() {
		fmt.Println(err.Error())
		return models.Budget{}
	}

	var res models.Budget

	rows.Scan(&res.BudgetID, &res.OwnerID, &res.BudgetName)

	return res
}
