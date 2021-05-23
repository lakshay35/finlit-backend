package expense

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	roleService "github.com/lakshay35/finlit-backend/services/role"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/shopspring/decimal"
)

// GetExpense ...
// Gets expense based on expense_id
func GetExpense(id uuid.UUID) (*models.Expense, error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT * FROM expenses WHERE expense_id = $1"

	stmt := database.PrepareStatement(connection, query)

	var expense models.Expense

	fmt.Println("Querying for expenses with id ", id)

	err := stmt.QueryRow(id).Scan(
		&expense.ExpenseID,
		&expense.BudgetID,
		&expense.ExpenseName,
		&expense.ExpenseValue,
		&expense.ExpenseDescription,
		&expense.ExpenseChargeCycle.ExpenseChargeCycleID,
	)

	if err != nil {
		panic(err)
	}

	unit, err := GetExpenseChargeCycleName(expense.ExpenseChargeCycle.ExpenseChargeCycleID)

	if err != nil {
		panic(err)
	}

	expense.ExpenseChargeCycle.Unit = unit

	if err != nil {
		return &models.Expense{}, err
	}

	return &expense, nil
}

// GetExpenseChargeCycleID ...
// Gets expense charge cycle id for a given expense name
func GetExpenseChargeCycleID(expenseName string) (int, error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT expense_charge_cycle_id from expense_charge_cycles WHERE unit = $1"

	stmt := database.PrepareStatement(connection, query)

	var expense_charge_cycle_id int

	err := stmt.QueryRow(expenseName).Scan(&expense_charge_cycle_id)

	if err != nil {
		return -1, err
	}

	return expense_charge_cycle_id, nil
}

// GetExpenseChargeCycleName ...
// Gets the name of an expense charge cycle based on id
func GetExpenseChargeCycleName(expenseChargeCycleID int) (string, error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT unit from expense_charge_cycles WHERE expense_charge_cycle_id = $1"

	stmt := database.PrepareStatement(connection, query)

	var unit string

	err := stmt.QueryRow(expenseChargeCycleID).Scan(&unit)

	if err != nil {
		return "", err
	}

	return unit, nil
}

// GetExpenseChargeCycles ...
// Gets all the types of expense charge cycles
func GetExpenseChargeCycles() []models.ExpenseChargeCycle {
	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := "SELECT * FROM expense_charge_cycles"

	rows, err := connection.Query(query)

	if err != nil {
		panic(err)
	}

	cycles := make([]models.ExpenseChargeCycle, 0)

	for rows.Next() {
		var cycle models.ExpenseChargeCycle

		err = rows.Scan(&cycle.ExpenseChargeCycleID, &cycle.Unit, &cycle.Days)

		if err != nil {
			panic(err)
		}

		cycles = append(cycles, cycle)
	}

	rows.Close()

	return cycles
}

// DeleteExpense ...
// Deletes expense based on id
func DeleteExpense(expenseID uuid.UUID, budgetID uuid.UUID, userID uuid.UUID) *errors.Error {

	// Ensure user is authorized to delete expense
	if !roleService.IsUserOwner(budgetID, userID) && !roleService.IsUserAdmin(budgetID, userID) {
		return &errors.Error{
			Message:    "You do not have enough permissions to delete expenses from this budget",
			StatusCode: http.StatusUnauthorized,
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := `DELETE FROM expenses WHERE expense_id = $1`

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(
		expenseID,
	)

	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// DeleteAllBudgetExpenses ...
// Deletes expense based on id
func DeleteAllBudgetExpenses(budgetID uuid.UUID, userID uuid.UUID) *errors.Error {

	// Ensure user is authorized to delete expense
	if !roleService.IsUserOwner(budgetID, userID) {
		return &errors.Error{
			Message:    "You do not have enough permissions to delete all expenses from this budget",
			StatusCode: http.StatusUnauthorized,
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := `DELETE FROM expenses WHERE budget_id = $1`

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(
		budgetID,
	)

	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// UpdateExpense ...
// Updates expense if user is owner or admin
func UpdateExpense(expense *models.Expense, userID uuid.UUID) *errors.Error {

	// Ensure user is authorized to update expense
	if !roleService.IsUserOwner(expense.BudgetID, userID) && !roleService.IsUserAdmin(expense.BudgetID, userID) {
		return &errors.Error{
			Message:    "You do not have enough permissions to add expenses to this budget",
			StatusCode: http.StatusUnauthorized,
		}
	}

	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := `UPDATE expenses SET budget_id = $1, expense_name = $2, expense_value = $3, expense_description = $4, expense_charge_cycle_id = (SELECT expense_charge_cycle_id FROM expense_charge_cycles where unit = $5), WHERE expense_id = $7`

	stmt := database.PrepareStatement(connection, query)

	_, err := stmt.Exec(
		expense.BudgetID,
		expense.ExpenseName,
		decimal.NewFromFloat32(expense.ExpenseValue).DivRound(decimal.NewFromInt(1), 2),
		expense.ExpenseDescription,
		expense.ExpenseChargeCycle.Unit,
		expense.ExpenseID,
	)

	if err != nil {
		fmt.Println(err.Error())
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// GetAllExpensesForBudget ...
// Gets all expenses for a specific budgetID
func GetAllExpensesForBudget(budgetID uuid.UUID, userID uuid.UUID) ([]models.Expense, *errors.Error) {
	if !roleService.IsUserAdmin(budgetID, userID) && !roleService.IsUserOwner(budgetID, userID) {
		return nil, &errors.Error{
			Message:    "You do not have enough permissions to view expenses of this budget",
			StatusCode: http.StatusUnauthorized,
		}
	}

	connection := database.GetConnection()
	defer database.CloseConnection(connection)

	query := `SELECT expense_id, budget_id, expense_name, expense_value, expense_description, unit, ecc.expense_charge_cycle_id,
 	ecc.days FROM expenses ep JOIN expense_charge_cycles ecc ON ecc.expense_charge_cycle_id = ep.expense_charge_cycle_id
	WHERE ep.budget_id = $1`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(budgetID)

	if err != nil {
		panic(err)
	}

	expenses := make([]models.Expense, 0)

	for rows.Next() {
		var expense models.Expense

		err = rows.Scan(
			&expense.ExpenseID,
			&expense.BudgetID,
			&expense.ExpenseName,
			&expense.ExpenseValue,
			&expense.ExpenseDescription,
			&expense.ExpenseChargeCycle.Unit,
			&expense.ExpenseChargeCycle.ExpenseChargeCycleID,
			&expense.ExpenseChargeCycle.Days,
		)

		if err != nil {
			panic(err)
		}

		expense.ExpenseTransactionCategories = getExpenseBudgetTransactionCategories(expense.ExpenseID)

		expenses = append(expenses, expense)
	}

	rows.Close()

	return expenses, nil
}

func getExpenseBudgetTransactionCategories(expenseID uuid.UUID) []string {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT category_name FROM budget_expense_transaction_categories betc join budget_transaction_categories btc ON betc.budget_transaction_category_id = btc.budget_transaction_category_id WHERE expense_id = $1"

	stmt := database.PrepareStatement(connection, query)

	res, catErr := stmt.Query(expenseID)

	if catErr != nil {
		panic(catErr)
	}

	var budgetExpenseTransactionCategories []string

	for res.Next() {
		var category_name string

		categoryScanErr := res.Scan(&category_name)

		if categoryScanErr != nil {
			panic(categoryScanErr)
		}

		budgetExpenseTransactionCategories = append(budgetExpenseTransactionCategories, category_name)
	}

	return budgetExpenseTransactionCategories
}

// AddExpenseToBudget ...
// Adds expense to budget
func AddExpenseToBudget(expense *models.AddExpensePayload, userID uuid.UUID) (*models.Expense, *errors.Error) {
	expenseChargeCycleID, err := GetExpenseChargeCycleID(expense.ExpenseChargeCycle.Unit)

	if err != nil {
		return nil, &errors.Error{
			Message:    "expense_charge_cycle " + expense.ExpenseChargeCycle.Unit + " is not valid",
			StatusCode: http.StatusBadRequest,
		}
	}

	if !roleService.IsUserAdmin(expense.BudgetID, userID) && !roleService.IsUserOwner(expense.BudgetID, userID) {
		return nil, &errors.Error{
			Message:    "You do not have enough permissions to add expenses to this budget",
			StatusCode: http.StatusUnauthorized,
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := `INSERT INTO expenses (budget_id, expense_name, expense_value, expense_description, expense_charge_cycle_id
	) VALUES ($1, $2, $3, $4, $5) RETURNING expense_id, (SELECT days from expense_charge_cycles where expense_charge_cycle_id = $5)`

	stmt := database.PrepareStatement(connection, query)

	var expenseResult = models.Expense{
		BudgetID:           expense.BudgetID,
		ExpenseName:        expense.ExpenseName,
		ExpenseValue:       expense.ExpenseValue,
		ExpenseDescription: expense.ExpenseDescription,
		ExpenseChargeCycle: expense.ExpenseChargeCycle,
	}

	dbError := stmt.QueryRow(
		expense.BudgetID,
		expense.ExpenseName,
		expense.ExpenseValue,
		expense.ExpenseDescription,
		expenseChargeCycleID,
	).Scan(&expenseResult.ExpenseID)

	if dbError != nil {
		fmt.Println(dbError.Error())
		panic("Something went wrong while adding expense")
	}

	// createBudgetExpenseTransactionCategoryMappings(expense.ExpenseTransactionCategories, expenseResult.ExpenseID, expense.BudgetID)
	query = "INSERT INTO budget_expense_transaction_categories (expense_id, budget_transaction_category_id) VALUES ($1, (Select budget_transaction_category_id from budget_transaction_categories WHERE category_name = $2 AND budget_id = $3))"

	stmt = database.PrepareStatement(connection, query)

	for _, cat := range expense.ExpenseTransactionCategories {
		_, err := stmt.Exec(expenseResult.ExpenseID, cat, expense.BudgetID)

		if err != nil {
			return nil, &errors.Error{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	expenseResult.ExpenseTransactionCategories = expense.ExpenseTransactionCategories

	return &expenseResult, nil
}
