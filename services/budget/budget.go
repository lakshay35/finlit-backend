package budget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	"github.com/lakshay35/finlit-backend/services/account"
	expenseService "github.com/lakshay35/finlit-backend/services/expense"
	roleService "github.com/lakshay35/finlit-backend/services/role"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
	"github.com/plaid/plaid-go/plaid"
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

// GetBudgetTransactionCategoryTransactions ...
// Gets budget transaction category transactions that the user has tagged
func GetBudgetTransactionCategoryTransactions(budgetID uuid.UUID) ([]models.BudgetTransactionCategoryTransaction, *errors.Error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT btct.transaction_name, btc.category_name FROM budget_transaction_category_transactions btct JOIN budget_transaction_categories btc on btc.budget_transaction_category_id = btct.budget_transaction_category_id WHERE btc.budget_id = $1"

	stmt := database.PrepareStatement(connection, query)

	res, queryErr := stmt.Query(budgetID)

	if queryErr != nil {
		return nil, &errors.Error{
			Message:    queryErr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	var transactions []models.BudgetTransactionCategoryTransaction

	for res.Next() {
		var temp models.BudgetTransactionCategoryTransaction

		scanErr := res.Scan(&temp.TransactionName, &temp.CategoryName)

		if scanErr != nil {
			panic(scanErr)
		}

		transactions = append(transactions, temp)
	}

	return transactions, nil
}

// GetBudgetExpenseSummary ...
// Calculates the budget expense summary for the past 30 day period
func GetBudgetExpenseSummary(budgetID uuid.UUID, userID uuid.UUID) (*models.ExpenseSummary, *errors.Error) {
	if !roleService.IsUserAdmin(budgetID, userID) && !roleService.IsUserOwner(budgetID, userID) {
		return nil, &errors.Error{
			Message:    "You are not entitled to this budget",
			StatusCode: http.StatusForbidden,
		}
	}

	budgetTransactionSources, getBudgetTransactionSourcesError := GetBudgetTransactionSources(budgetID)

	if getBudgetTransactionSourcesError != nil {
		return nil, getBudgetTransactionSourcesError
	}

	expenses, expensesErr := expenseService.GetAllExpensesForBudget(budgetID, userID)

	if expensesErr != nil {
		return nil, expensesErr
	}

	var txs = make([]plaid.Transaction, 0)

	for _, bts := range budgetTransactionSources {
		transactions, getTransactionsErr := account.GetTransactions(bts.ExternalAccountID, time.Now().Local().Add(-30*24*time.Hour).Format("2006-01-02"),
			time.Now().Local().Format("2006-01-02"))

		if getTransactionsErr != nil {
			return nil, getTransactionsErr
		}

		txs = append(txs, transactions...)
	}

	budgetTransactionCategoryTransactions, budgetTransactionCategoryTransactionsErr := GetBudgetTransactionCategoryTransactions(budgetID)

	if budgetTransactionCategoryTransactionsErr != nil {
		return nil, budgetTransactionCategoryTransactionsErr
	}

	summary := calculatedBudgetExpenseSummaryUsingTransactionsAndExpenses(expenses, txs, budgetTransactionCategoryTransactions)

	return &summary, nil
	// plaidSerivce.PlaidClient().

	// return "hello", nil
}

// GetAllBudgetTransactionCategories ...
// Gets all budget transaction categories
func GetAllBudgetTransactionCategories(budgetID uuid.UUID) ([]models.BudgetTransactionCategory, *errors.Error) {
	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := "SELECT * FROM budget_transaction_categories WHERE budget_id = $1"

	stmt := database.PrepareStatement(connection, query)

	res, execErr := stmt.Query(budgetID)

	if execErr != nil {
		return nil, &errors.Error{
			StatusCode: http.StatusBadRequest,
			Message:    execErr.Error(),
		}
	}

	categories := make([]models.BudgetTransactionCategory, 0)

	for res.Next() {
		var temp models.BudgetTransactionCategory

		scanErr := res.Scan(&temp.BudgetTransactionCategoryID, &temp.BudgetID, &temp.CategoryName)

		if scanErr != nil {
			panic(scanErr)
		}

		categories = append(categories, temp)
	}

	return categories, nil
}

func calculatedBudgetExpenseSummaryUsingTransactionsAndExpenses(
	expenses []models.Expense,
	transactions []plaid.Transaction,
	categories []models.BudgetTransactionCategoryTransaction,
) models.ExpenseSummary {
	summary := models.ExpenseSummary{}

	transactionCategoriesMap := make(map[string]string)
	for _, cat := range categories {
		transactionCategoriesMap[cat.TransactionName] = cat.CategoryName
	}

	res := make(map[string]models.ExpenseCategorySummary)
	res["Uncategorized"] = models.ExpenseCategorySummary{
		CategoryName:   "Uncategorized",
		ExpenseLimit:   0.0,
		CurrentExpense: 0.0,
		Transactions:   make([]plaid.Transaction, 0),
	}

	for _, tx := range transactions {
		if tx.Amount > 0 && tx.Category[0] != "Payment" && tx.Category[0] != "Transfer" {

			var temp models.ExpenseCategorySummary

			transactionCategory := transactionCategoriesMap[tx.Name]

			if transactionCategory != "" {
				if res[transactionCategory].CategoryName != "" {
					temp := res[transactionCategory]
					temp.CurrentExpense = temp.CurrentExpense + tx.Amount
					temp.Transactions = append(temp.Transactions, tx)
					res[transactionCategory] = temp
				} else {
					temp.CategoryName = transactionCategory
					txs := make([]plaid.Transaction, 0)
					temp.Transactions = append(txs, tx)
					res[transactionCategory] = temp
				}
			} else {
				temp := res["Uncategorized"]
				temp.CurrentExpense += tx.Amount
				temp.Transactions = append(temp.Transactions, tx)
				res["Uncategorized"] = temp
			}
		}
	}

	// other.CurrentExpense = value

	arr := make([]models.ExpenseCategorySummary, 0)
	for _, v := range res {
		arr = append(arr, v)
	}

	summary.ExpenseCategories = arr

	return summary
}

// GetTransactionCategories ...
// Get transaction categories for a given budget
func GetTransactionCategories(budgetID uuid.UUID) ([]models.BudgetTransactionCategory, *errors.Error) {

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT * FROM budget_transaction_categories WHERE budget_id = $1"

	stmt := database.PrepareStatement(connection, query)

	res, queryErr := stmt.Query(budgetID)

	if queryErr != nil {
		return nil, &errors.Error{
			Message:    queryErr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	categories := make([]models.BudgetTransactionCategory, 0)

	for res.Next() {
		var temp models.BudgetTransactionCategory

		scanErr := res.Scan(&temp.BudgetTransactionCategoryID, &temp.BudgetID, &temp.CategoryName)

		if scanErr != nil {
			panic(scanErr)
		}

		categories = append(categories, temp)
	}

	return categories, nil
}

// DeleteTransactionCategoryTransactions ...
func DeleteTransactionCategoryTransactions(transactionCategoryID uuid.UUID) *errors.Error {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	// TODO: Delete all transaction associations with category (CASCADE ACTION)
	query := "DELETE FROM budget_transaction_category_transactions where budget_transaction_category_id = $1"

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(transactionCategoryID)

	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// DeleteBudgetTransactionCategoryExpense ...
func DeleteBudgetTransactionCategoryExpense(transactionCategoryID uuid.UUID) *errors.Error {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	// TODO: Delete all transaction associations with category (CASCADE ACTION)
	query := "DELETE FROM expenses where budget_transaction_category_id = $1"

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(transactionCategoryID)

	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// DeleteTransactionCategory ...
// Deletes transaction category
func DeleteTransactionCategory(categoryID uuid.UUID) *errors.Error {

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	deleteExpenseErr := DeleteBudgetTransactionCategoryExpense(categoryID)

	if deleteExpenseErr != nil {
		return deleteExpenseErr
	}

	deleteTransactionCategoryTransactionsErr := DeleteTransactionCategoryTransactions(categoryID)

	if deleteTransactionCategoryTransactionsErr != nil {
		return deleteTransactionCategoryTransactionsErr
	}

	query := "DELETE FROM budget_transaction_categories where budget_transaction_category_id = $1"

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(categoryID)

	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// CreateTransactionCategory ...
// Creates a transaction category for the given budget
func CreateTransactionCategory(category models.BudgetTransactionCategoryCreationPayload, userID uuid.UUID) (*models.BudgetTransactionCategory, *errors.Error) {

	if !roleService.IsUserAdmin(category.BudgetID, userID) && !roleService.IsUserOwner(category.BudgetID, userID) {
		return nil, &errors.Error{
			StatusCode: http.StatusForbidden,
			Message:    "You are not authorized to create a transaction category for the given budget",
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "INSERT INTO budget_transaction_categories (budget_id, category_name) VALUES ($1, $2) RETURNING budget_transaction_category_id, budget_id, category_name"

	stmt := database.PrepareStatement(connection, query)

	var temp models.BudgetTransactionCategory

	scanErr := stmt.QueryRow(category.BudgetID, category.CategoryName).Scan(&temp.BudgetTransactionCategoryID, &temp.BudgetID, &temp.CategoryName)

	if scanErr != nil {
		return nil, &errors.Error{
			Message:    scanErr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return &temp, nil
}
