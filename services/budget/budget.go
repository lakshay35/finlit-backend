package budget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	expenseService "github.com/lakshay35/finlit-backend/services/expense"
	roleService "github.com/lakshay35/finlit-backend/services/role"
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

// GetBudgetTransactionSources ...
// Retrieves a list of all budget transaction sources
func GetBudgetTransactionSources(budgetID uuid.UUID) ([]models.BudgetTransactionSourcePayload, *errors.Error) {
	connection := database.GetConnection()

	query := "SELECT ea.external_account_id, ea.account_name, bts.budget_id, bts.budget_transaction_source_id FROM external_accounts ea JOIN budget_transaction_sources bts ON bts.external_account_id = ea.external_account_id WHERE bts.budget_id = $1"

	stmt := database.PrepareStatement(connection, query)

	rows, dbError := stmt.Query(budgetID)

	if dbError != nil {
		return nil, &errors.Error{
			StatusCode: 400,
			Message:    dbError.Error(),
		}
	}

	accounts := make([]models.BudgetTransactionSourcePayload, 0)

	for rows.Next() {
		var temp models.BudgetTransactionSourcePayload
		scanErr := rows.Scan(&temp.ExternalAccountID, &temp.AccountName, &temp.BudgetID, &temp.BudgetTransactionSourceID)

		if scanErr != nil {
			panic(scanErr)
		}

		accounts = append(accounts, temp)
	}

	return accounts, nil
}

// CreateBudgetTransactionSource ...
// Creates a budget transaction source
func CreateBudgetTransactionSource(budgetTransactionSource models.BudgetTransactionSourceCreationPayload) (*models.BudgetTransactionSource, *errors.Error) {
	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	// TODO: Refactor into stored procedure call

	query := "INSERT INTO budget_transaction_sources (external_account_id, budget_id) VALUES ($1, $2) RETURNING budget_transaction_source_id"

	stmt := database.PrepareStatement(connection, query)

	var res models.BudgetTransactionSource

	res.BudgetID = budgetTransactionSource.BudgetID
	res.ExternalAccountID = budgetTransactionSource.ExternalAccountID

	dbError := stmt.QueryRow(budgetTransactionSource.ExternalAccountID, budgetTransactionSource.BudgetID).Scan(&res.BudgetTransactionSourceID)

	if dbError != nil {
		return nil, &errors.Error{
			StatusCode: 400,
			Message:    dbError.Error(),
		}
	}

	return &res, nil
}

// DeleteBudgetTransactionSource ...
// Delete budget transation source
func DeleteBudgetTransactionSource(budgetTransactionSourceID uuid.UUID, userID uuid.UUID) *errors.Error {

	budgetTransactionSource, getBudgetTransactionSourceError := GetBudgetTransactionSource(budgetTransactionSourceID)

	if getBudgetTransactionSourceError != nil {
		return getBudgetTransactionSourceError
	}

	if roleService.IsUserAdmin(budgetTransactionSource.BudgetID, userID) || roleService.IsUserOwner(budgetTransactionSource.BudgetID, userID) {

		connection := database.GetConnection()

		defer database.CloseConnection(connection)

		query := "DELETE FROM budget_transaction_sources WHERE budget_transaction_source_id = $1"

		stmt := database.PrepareStatement(connection, query)

		_, dbError := stmt.Exec(budgetTransactionSourceID)

		if dbError != nil {
			return &errors.Error{
				StatusCode: 400,
				Message:    dbError.Error(),
			}
		}

		return nil
	}

	return &errors.Error{
		StatusCode: http.StatusForbidden,
		Message:    "You are not authorized to delete this transaction source",
	}
}

// GetBudgetTransactionSource ...
// Gets budget transaction source by id
func GetBudgetTransactionSource(budgetTransactionSourceID uuid.UUID) (*models.BudgetTransactionSource, *errors.Error) {
	connection := database.GetConnection()

	query := "SELECT * FROM budget_transaction_sources WHERE budget_transaction_source_id = $1"

	stmt := database.PrepareStatement(connection, query)

	var res models.BudgetTransactionSource
	err := stmt.QueryRow(budgetTransactionSourceID).Scan(&res.BudgetTransactionSourceID, &res.ExternalAccountID, &res.BudgetID)

	if err != nil {
		fmt.Println("no budget exists for id", budgetTransactionSourceID)
		return nil, &errors.Error{
			Message:    "No budget transaction source exists with provided id",
			StatusCode: http.StatusNotFound,
		}
	}

	return &res, nil
}

// DoesBudgetExist ...
// Checks if a budget exists
func DoesBudgetExist(UserID uuid.UUID, budgetName string) bool {
	connection := database.GetConnection()

	query := "SELECT * FROM budgets WHERE owner_id = $1 AND budget_name = $2"

	stmt, err := connection.Prepare(query)

	if err != nil {
		return false
	}

	res, err := stmt.Query(UserID, budgetName)

	if err != nil {
		panic(err.Error())
	}

	defer res.Close()
	defer database.CloseConnection(connection)

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

	rows.Close()

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

	if budgetName == "" {
		return nil, &errors.Error{
			Message:    "Budget name need to be a non-empty string",
			StatusCode: http.StatusBadRequest,
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

	fmt.Println(result)

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

	res.Close()

	return result, nil
}

// DeleteAllBudgetTransactionSources ...
// Deletes all budget transaction sources
func DeleteAllBudgetTransactionSources(budgetID uuid.UUID) *errors.Error {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "DELETE FROM budget_transaction_sources WHERE budget_id = $1"

	stmt := database.PrepareStatement(connection, query)

	_, stmtErr := stmt.Exec(budgetID)

	if stmtErr != nil {
		return &errors.Error{
			Message:    "Buget Transaction Sources not found",
			StatusCode: http.StatusNotFound,
		}
	}

	return nil
}

// DeleteBudget ...
// Deletes budget an all associated expenses
func DeleteBudget(budgetID uuid.UUID, userID uuid.UUID) *errors.Error {
	if !roleService.IsUserOwner(budgetID, userID) {
		return &errors.Error{
			Message:    "User requesting deletion needs to be the owner of budget to proceed",
			StatusCode: http.StatusUnauthorized,
		}
	}

	deleteBTSErr := DeleteAllBudgetTransactionSources(budgetID)

	if deleteBTSErr != nil {
		return deleteBTSErr
	}

	err := expenseService.DeleteAllBudgetExpenses(budgetID, userID)

	if err != nil {
		return err
	}

	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := "DELETE FROM budgets where budget_id = $1"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Error preparing query for deleting budgets")
	}

	_, errrr := stmt.Exec(budgetID)

	if errrr != nil {
		return &errors.Error{
			Message:    errr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}
