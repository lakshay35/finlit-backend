package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// AddExpense ..
// @Summary Adds an expense to the database
// @Description Add expense to an existing budget
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Expense
// @Router /expense/add [post]
func AddExpense(c *gin.Context) {

	var expense models.Expense
	err := c.BindJSON(&expense)

	if err != nil {
		requests.ThrowError(c, http.StatusBadRequest, "Payload does not match")
	}

	expenseChargeCycleID, err := getExpenseID(expense.ExpenseChargeCycle)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"expense_charge_cycle "+expense.ExpenseChargeCycle+" is not valid",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	if !isUserAdmin(expense.BudgetID, user.UserID) && !isUserOwner(expense.BudgetID, user.UserID) {
		requests.ThrowError(
			c,
			http.StatusUnauthorized,
			"You do not have enough permissions to add expenses to this budget",
		)
		return
	}

	connection := database.GetConnection()

	defer connection.Commit()

	query := `INSERT INTO expenses (budget_id, expense_name, expense_value, expense_description, expense_charge_cycle_id
	) VALUES ($1, $2, $3, $4, $5) RETURNING expense_id`

	stmt := database.PrepareStatement(connection, query)

	var expense_id uuid.UUID

	err = stmt.QueryRow(
		expense.BudgetID,
		expense.ExpenseName,
		expense.ExpenseValue,
		expense.ExpenseDescription,
		expenseChargeCycleID,
	).Scan(&expense_id)

	if err != nil {
		panic("Something went wrong while adding expense")
	}

	expense.ExpenseID = expense_id

	c.JSON(
		http.StatusCreated,
		gin.H{
			"message": "Successfully created expense",
			"data":    expense,
		},
	)
}

// GetAllExpenses ..
// @Summary Gets expenses for budget
// @Description Gets a list of all expenses tied to a given budget
// @Accept  json
// @Produce  json
// @Success 200 {object} []models.Expense
// @Router /expense/add [post]
func GetAllExpenses(c *gin.Context) {

	var budget models.Budget
	c.BindJSON(&budget)

	user := requests.GetUserFromContext(c)

	if !isUserViewer(budget.BudgetID, user.UserID) && !isUserOwner(budget.BudgetID, user.UserID) {
		requests.ThrowError(
			c,
			http.StatusUnauthorized,
			"You do not have enough permissions to view expenses of this budget",
		)
		return
	}

	connection := database.GetConnection()

	defer connection.Commit()

	query := `SELECT expense_id, budget_id, expense_name, expense_value, expense_description, unit
	FROM expenses ep JOIN expense_charge_cycles ecc ON ecc.expense_charge_cycle_id = ep.expense_charge_cycle_id
	WHERE ep.budget_id = $1`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(budget.BudgetID)

	if err != nil {
		panic(err)
	}

	expenses := make([]models.Expense, 0)

	for rows.Next() {
		var expense models.Expense
		var unit string

		rows.Scan(
			&expense.ExpenseID,
			&expense.BudgetID,
			&expense.ExpenseName,
			&expense.ExpenseValue,
			&expense.ExpenseDescription,
			&unit,
		)

		expense.ExpenseChargeCycle = unit

		expenses = append(expenses, expense)
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "Successfully retrieved all expenses",
			"data":    expenses,
		},
	)

}

// TODO: Check permissions and authorizations
// UpdateExpense ..
// @Summary Adds an expense to the database
// @Description Add expense to an existing budget
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Expense
// @Router /expense/add [post]
func UpdateExpense(c *gin.Context) {

	var expense models.Expense

	err := requests.ParseBody(c, &expense)

	if err != nil {
		return
	}

	if !doesExpenseExist(expense.ExpenseID) {
		requests.ThrowError(
			c,
			http.StatusNotFound,
			"Request contains an expense_id that does not exist",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	// expenseID, err := getExpenseID(expense.ExpenseChargeCycle)

	if err != nil {
		// requests.ThrowError(
		// 	c,
		// 	http.StatusBadRequest,
		// 	"expense_charge_cycle %s is not valid", expense.ExpenseChargeCycle,
		// )
		return
	}
	isUserAdmin := isUserAdmin(expense.BudgetID, user.UserID)
	isUserOwner := isUserOwner(expense.BudgetID, user.UserID)
	fmt.Printf("Admin: %t Owner: %t", isUserAdmin, isUserOwner)
	if !isUserAdmin && !isUserOwner {
		requests.ThrowError(
			c,
			http.StatusUnauthorized,
			"You do not have enough permissions to add expenses to this budget",
		)
		return
	}

	connection := database.GetConnection()

	defer connection.Commit()

	query := `UPDATE expenses (budget_id, expense_name, expense_value, expense_description, expense_charge_cycle_id
	) VALUES ($1, $2, $3, $4, $5) WHERE expense_id = $6`

	stmt := database.PrepareStatement(connection, query)

	_, err = stmt.Exec(
		expense.BudgetID,
		expense.ExpenseName,
		expense.ExpenseValue,
		expense.ExpenseDescription,
		expense.ExpenseChargeCycle,
		expense.ExpenseID,
	)

	if err != nil {
		panic("Something went wrong while adding expense")
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"message": "Successfully updated expense",
			"data":    expense,
		},
	)
}

// TODO: Check permission and authorization
// DeleteExpense ..
// @Summary Adds an expense to the database
// @Description Deletes Expense
// @Accept  json
// @Param id path string true "Expense ID (UUID)"
// @Produce  json
// @Success 200 {object} models.Expense
// @Router /expense/add [post]
func DeleteExpense(c *gin.Context) {

	var expense models.Expense

	err := requests.ParseBody(c, &expense)

	if err != nil {
		return
	}

	connection := database.GetConnection()

	defer connection.Commit()

	query := `UPDATE expenses (budget_id, expense_name, expense_value, expense_description, expense_charge_cycle_id
	) VALUES ($1, $2, $3, $4, $5) WHERE expense_id = $6`

	stmt := database.PrepareStatement(connection, query)

	_, err = stmt.Exec(
		expense.BudgetID,
		expense.ExpenseName,
		expense.ExpenseValue,
		expense.ExpenseDescription,
		expense.ExpenseChargeCycle,
		expense.ExpenseID,
	)

	if err != nil {
		panic("Something went wrong while adding expense")
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"message": "Successfully updated expense",
			"data":    expense,
		},
	)
}

// GetExpenseChargeCycles ...
// Gets all the expense charge cycles
// available to create an expense for a budget
func GetExpenseChargeCycles(c *gin.Context) {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT * FROM expense_charge_cycles"

	rows, err := connection.Query(query)

	if err != nil {
		panic(err)
	}

	cycles := make([]models.ExpenseChargeCycle, 0)

	for rows.Next() {
		var cycle models.ExpenseChargeCycle

		rows.Scan(&cycle.ExpenseChargeCycleID, &cycle.Unit)

		cycles = append(cycles, cycle)
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "Successfully retrieved all expense charge cycles",
			"data":    cycles,
		},
	)

}

// doesExpenseExist
// Determines if their is an expense with the given ID
func doesExpenseExist(id uuid.UUID) bool {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT * FROM expenses WHERE expense_id = $1"

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(id)

	if err != nil {
		panic(err)
	}

	return rows.Next()
}

// isUserAdmin
// Determines if user is an admin on the given budget
func isUserAdmin(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	defer connection.Commit()

	query := `SELECT * FROM user_roles ur WHERE ur.user_id = $1 AND ur.role_id = (SELECT role_id FROM roles WHERE role_name = 'Full Rights')
	AND ur.budget_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		panic(err)
	}

	return rows.Next()
}

// isUserAdmin
// Determines if user is an admin on the given budget
func isUserViewer(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT * FROM user_roles ur WHERE ur.user_id = $1 AND ur.budget_id = $2"

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		panic(err)
	}

	return rows.Next()
}

// isUserOwner
// Checks if user is an owner
func isUserOwner(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	defer connection.Commit()

	query := `SELECT owner_id FROM budgets WHERE budget_id = $1 AND owner_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(budgetID, userID)

	if err != nil {
		panic(err)
	}

	return rows.Next()
}

// getExpenseID
// Gets expense id for a given expense name
func getExpenseID(expenseName string) (int, error) {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT expense_charge_cycle_id from expense_charge_cycles WHERE unit = $1"

	stmt := database.PrepareStatement(connection, query)

	var expense_charge_cycle_id int

	err := stmt.QueryRow(expenseName).Scan(&expense_charge_cycle_id)

	if err != nil {
		return -1, err
	}

	return expense_charge_cycle_id, nil
}
