package budget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"

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
	defer database.CloseConnection(connection)

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
	defer database.CloseConnection(connection)

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

	err = rows.Scan(&res.BudgetID, &res.OwnerID, &res.BudgetName)

	if err != nil {
		panic(err)
	}

	return res
}

// CreateBudget ...
// Creates budget if it doesn't already exist for user
func CreateBudget(userID uuid.UUID, budgetName string) (*models.Budget, *errors.Error) {
	if DoesBudgetExist(userID, budgetName) {
		return nil, &errors.Error{
			Message:    "Budget named " + budgetName + " already exists",
			StatusCode: http.StatusConflict,
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "INSERT INTO budgets (owner_id, budget_name) VALUES ($1, $2) RETURNING owner_id, budget_name, budget_id"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Something went wrong when preparing query to create budget")
	}

	var result models.Budget

	errr = stmt.QueryRow(userID, budgetName).Scan(
		&result.OwnerID,
		&result.BudgetName,
		&result.BudgetID,
	)

	if errr != nil {
		panic(errr)
	}

	return &result, nil
}

// GetAllBudgets ...
// Gets all budgets that given userID owns
// TODO: Get all budgets user owns and has access to, include access type in return object
func GetAllBudgets(userID uuid.UUID) ([]models.Budget, *errors.Error) {
	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := "SELECT * FROM budgets where owner_id = $1"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Error preparing query for getting budgets")
	}

	res, errr := stmt.Query(userID)

	if errr != nil {
		return nil, &errors.Error{
			Message:    "No Budgets Found",
			StatusCode: http.StatusNotFound,
		}
	}

	var result []models.Budget = make([]models.Budget, 0)

	for res.Next() {
		var temp models.Budget
		err := res.Scan(&temp.BudgetID, &temp.BudgetName, &temp.OwnerID)

		if err != nil {
			panic(err)
		}

		// Appends the item to the result
		result = append(result, temp)
	}

	return result, nil
}
