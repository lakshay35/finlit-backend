package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	expenseService "github.com/lakshay35/finlit-backend/services/expense"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// AddExpense ..
// @Summary Adds an expense to the given budget
// @Description Add expense to an existing budget
// @Tags Budget Expenses
// @Accept  json
// @Produce  json
// @Param body body models.Expense true "Expense payload representing entity to be created"
// @Security Google AccessToken
// @Success 201 {object} models.Expense
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 401 {object} models.Error
// @Router /expense/add [post]
func AddExpense(c *gin.Context) {

	var json models.Expense
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	user := requests.GetUserFromContext(c)

	expense, err := expenseService.AddExpenseToBudget(&json, user.UserID)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

		return
	}
	c.JSON(http.StatusCreated, expense)
}

// GetAllExpenses ..
// @Summary Gets expenses for budget
// @Description Gets a list of all expenses tied to a given budget
// @Tags Budget Expenses
// @Accept  json
// @Produce  json
// @Param Budget-ID header string true "Budget ID to get expenses against"
// @Security Google AccessToken
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

	expenses, getExpensesError := expenseService.GetAllExpensesForBudget(budgetID, user.UserID)

	if getExpensesError != nil {
		requests.ThrowError(
			c,
			getExpensesError.StatusCode,
			getExpensesError.Error(),
		)
	}

	c.JSON(http.StatusOK, expenses)

}

// UpdateExpense ..
// @Summary Adds an expense to the database
// @Description Add expense to an existing budget
// @Tags Budget Expenses
// @Accept  json
// @Produce  json
// @Param body body models.Expense true "Expense payload representing entity to be updated"
// @Security Google AccessToken
// @Success 204 {object} models.Expense
// @Failure 403 {object} models.Error
// @Failure 401 {object} models.Error
// @Failure 404 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /expense/update [put]
func UpdateExpense(c *gin.Context) {

	var json models.Expense

	err := requests.ParseBody(c, &json)
	if err != nil {
		return
	}

	// Ensure expense exists
	_, err = expenseService.GetExpense(json.ExpenseID)
	if err != nil {
		requests.ThrowError(
			c,
			http.StatusNotFound,
			"Request contains an expense_id that does not exist",
		)
		return
	}

	user := requests.GetUserFromContext(c)

	updateExpenseError := expenseService.UpdateExpense(
		&json,
		user.UserID,
	)

	if updateExpenseError != nil {
		requests.ThrowError(
			c,
			updateExpenseError.StatusCode,
			updateExpenseError.Message,
		)

		return
	}

	c.JSON(http.StatusNoContent, json)
}

// DeleteExpense ..
// @Summary Deletes Expense
// @Description Deletes Expense from DB based on id
// @Tags Budget Expenses
// @Accept  json
// @Param id path string true "Expense ID (UUID)"
// @Produce  json
// @Security Google AccessToken
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

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusNotFound,
			err.Error(),
		)
	}

	user := requests.GetUserFromContext(c)

	deleteExpenseError := expenseService.DeleteExpense(id, expense.BudgetID, user.UserID)

	if deleteExpenseError != nil {
		requests.ThrowError(
			c,
			deleteExpenseError.StatusCode,
			deleteExpenseError.Error(),
		)

		return
	}

	c.Status(http.StatusNoContent)
}

// GetExpenseChargeCycles ...
// @Summary Gets a list of expense charge cycles
// @Description Gets all the expense charge cycles available to create an expense for a budget
// @Tags Budget Expenses
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {array} models.ExpenseChargeCycle
// @Failure 403 {object} models.Error
// @Router /expense/get-expense-charge-cycles [get]
func GetExpenseChargeCycles(c *gin.Context) {
	c.JSON(http.StatusOK, expenseService.GetExpenseChargeCycles())
}
