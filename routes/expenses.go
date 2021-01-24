package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	expenseService "github.com/lakshay35/finlit-backend/services/expense"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// AddExpense ..
// @Summary Adds an expense to the given budget
// @Description Add expense to an existing budget
// @Tags expense
// @Accept  json
// @Produce  json
// @Param body body models.Expense true "Expense payload representing entity to be created"
// @Security ApiKeyAuth
// @Success 201 {object} models.Expense
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 401 {object} models.Error
// @Router /expense/add [post]
func AddExpense(c *gin.Context) {

	var expense models.Expense
	err := c.BindJSON(&expense)

	if err != nil {
		requests.ThrowError(c, http.StatusBadRequest, "Payload does not match")
	}

	expenseChargeCycleID, err := expenseService.GetExpenseChargeCycleID(expense.ExpenseChargeCycle)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"expense_charge_cycle "+expense.ExpenseChargeCycle+" is not valid",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	if !expenseService.IsUserAdmin(expense.BudgetID, user.UserID) && !expenseService.IsUserOwner(expense.BudgetID, user.UserID) {
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

	err = stmt.QueryRow(
		expense.BudgetID,
		expense.ExpenseName,
		expense.ExpenseValue,
		expense.ExpenseDescription,
		expenseChargeCycleID,
	).Scan(&expense.ExpenseID)

	if err != nil {
		panic("Something went wrong while adding expense")
	}

	c.JSON(http.StatusCreated, expense)
}

// GetAllExpenses ..
// @Summary Gets expenses for budget
// @Description Gets a list of all expenses tied to a given budget
// @Tags expense
// @Accept  json
// @Produce  json
// @Param budgetID header string true "Budget ID to get expenses against"
// @Security ApiKeyAuth
// @Success 200 {array} models.Expense
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 401 {object} models.Error
// @Router /expense/get [get]
func GetAllExpenses(c *gin.Context) {

	budgetID, err := uuid.Parse(c.GetHeader("Budget-ID"))

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Header 'Budget-ID' missing in request",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	if !expenseService.IsUserAdmin(budgetID, user.UserID) && !expenseService.IsUserOwner(budgetID, user.UserID) {
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

	rows, err := stmt.Query(budgetID)

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

	c.JSON(http.StatusOK, expenses)

}

// UpdateExpense ..
// @Summary Adds an expense to the database
// @Description Add expense to an existing budget
// @Tags expense
// @Accept  json
// @Produce  json
// @Param body body models.Expense true "Expense payload representing entity to be updated"
// @Security ApiKeyAuth
// @Success 204 {object} models.Expense
// @Failure 403 {object} models.Error
// @Failure 401 {object} models.Error
// @Failure 404 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /expense/update [put]
func UpdateExpense(c *gin.Context) {

	var expense models.Expense

	err := requests.ParseBody(c, &expense)
	if err != nil {
		return
	}

	// Ensure expense exists
	_, err = expenseService.GetExpense(expense.ExpenseID)
	if err != nil {
		requests.ThrowError(
			c,
			http.StatusNotFound,
			"Request contains an expense_id that does not exist",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	// Ensure user is authorized to update expense
	if !expenseService.IsUserOwner(expense.BudgetID, user.UserID) && !expenseService.IsUserAdmin(expense.BudgetID, user.UserID) {
		requests.ThrowError(
			c,
			http.StatusUnauthorized,
			"You do not have enough permissions to add expenses to this budget",
		)
		return
	}

	// Ensures expenseID is passed
	if expense.ExpenseID.String() != "" {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Valid expense_id not passed",
		)
	}

	connection := database.GetConnection()
	defer connection.Commit()

	query := `UPDATE expenses (budget_id, expense_name, expense_value, expense_description, expense_charge_cycle_id
	) VALUES ($1, $2, $3, $4, $5) WHERE expense_id = $6 `

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
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
	}

	c.JSON(http.StatusNoContent, expense)
}

// DeleteExpense ..
// @Summary Deletes Expense
// @Description Deletes Expense from DB based on id
// @Tags expense
// @Accept  json
// @Param id path string true "Expense ID (UUID)"
// @Produce  json
// @Param id path string true "ID of Expense"
// @Security ApiKeyAuth
// @Success 204
// @Failure 403 {object} models.Error
// @Failure 401 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /expense/delete/{id} [delete]
func DeleteExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

		return
	}

	expense, err := expenseService.GetExpense(id)

	user := requests.GetUserFromContext(c)

	// Ensure user is authorized to update expense
	if !expenseService.IsUserOwner(expense.BudgetID, user.UserID) && !expenseService.IsUserOwner(expense.BudgetID, user.UserID) {
		requests.ThrowError(
			c,
			http.StatusUnauthorized,
			"You do not have enough permissions to delete expenses from this budget",
		)
		return
	}

	connection := database.GetConnection()

	defer connection.Commit()

	query := `DELETE FROM expenses WHERE expense_id = $1`

	stmt := database.PrepareStatement(connection, query)

	_, err = stmt.Exec(
		id,
	)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
	}
	c.Status(http.StatusNoContent)
}

// GetExpenseChargeCycles ...
// @Summary Gets a list of expense charge cycles
// @Description Gets all the expense charge cycles available to create an expense for a budget
// @Tags expense
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} models.ExpenseChargeCycle
// @Failure 403 {object} models.Error
// @Router /expense/get-expense-charge-cycles [get]
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

	c.JSON(http.StatusOK, cycles)
}
