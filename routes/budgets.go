package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	budgetService "github.com/lakshay35/finlit-backend/services/budget"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// CreateBudget ...
// @Summary Create a budget
// @Description Creates a budget with requesting user as owner
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param budget body models.CreateBudgetPayload true "Budget body needed to create budget"
// @Security Google AccessToken
// @Success 200 {object} models.Budget
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 409 {object} models.Error
// @Router /budget/create [post]
func CreateBudget(c *gin.Context) {
	var json models.CreateBudgetPayload
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(err)
	}

	budget, budgetCreationError := budgetService.CreateBudget(user.UserID, json.BudgetName)

	if budgetCreationError != nil {
		requests.ThrowError(
			c,
			budgetCreationError.StatusCode,
			budgetCreationError.Error(),
		)

		return
	}

	c.JSON(201, budget)
}

// GetBudgets ...
// @Summary Get Budgets
// @Description Gets a list of all budgets current user is a part of
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {array} models.Budget
// @Failure 403 {object} models.Error
// @Router /budget/get [get]
func GetBudgets(c *gin.Context) {

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	result, getAllBudgetsError := budgetService.GetAllBudgets(user.UserID)

	if getAllBudgetsError != nil {
		requests.ThrowError(
			c,
			getAllBudgetsError.StatusCode,
			getAllBudgetsError.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, result)
}

// GetBudgetTransactionSources ...
// @Summary Get Budget Transaction Sources
// @Description Gets a list of all budget transaction sources current user is a part of
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param Budget-ID header string true "Budget ID to pull transaction sources for"
// @Security Google AccessToken
// @Success 200 {array} models.BudgetTransactionSourcePayload
// @Failure 403 {object} models.Error
// @Router /budget/get-transaction-sources [get]
func GetBudgetTransactionSources(c *gin.Context) {

	budgetID, budgetIDError := uuid.Parse(c.GetHeader("Budget-ID"))

	if budgetIDError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Header 'Budget-ID' must contain a valid uuid",
		)

		return
	}

	result, err := budgetService.GetBudgetTransactionSources(budgetID)

	if err != nil {
		requests.ThrowError(
			c,
			err.StatusCode,
			err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateBudgetTransactionSource ...
// @Summary Creates a budget transaction source
// @Description Creates a budget transaction source
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param budgetTransactionSource body models.BudgetTransactionSourceCreationPayload true "Budget Transaction Source"
// @Security Google AccessToken
// @Success 200 {object} models.BudgetTransactionSource
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /budget/create-transaction-source [post]
func CreateBudgetTransactionSource(c *gin.Context) {
	var json models.BudgetTransactionSourceCreationPayload
	parseErr := requests.ParseBody(c, &json)

	if parseErr != nil {
		return
	}

	res, createBudgetTransactionError := budgetService.CreateBudgetTransactionSource(json)

	if createBudgetTransactionError != nil {
		requests.ThrowError(
			c,
			createBudgetTransactionError.StatusCode,
			createBudgetTransactionError.Message,
		)

		return
	}

	c.JSON(http.StatusOK, res)
}

// DeleteBudgetTransactionSource ...
// @Summary Delete Budget Transaction Source
// @Description Deletes a budget transaction source
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param budget-transaction-source-id path string true "Budget Transaction Source Id"
// @Security Google AccessToken
// @Success 204
// @Failure 404 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 403 {object} models.Error
// @Router /budget/delete-transaction-source/{budget-transaction-source-id} [delete]
func DeleteBudgetTransactionSource(c *gin.Context) {
	param := c.Param("budget-transaction-source-id")
	budgetTransactionSourceID, parseErr := uuid.Parse(param)

	if parseErr != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Budget Transaction Source ID must be a UUID",
		)

		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	deleteBudgetError := budgetService.DeleteBudgetTransactionSource(budgetTransactionSourceID, user.UserID)

	if deleteBudgetError != nil {
		requests.ThrowError(
			c,
			deleteBudgetError.StatusCode,
			deleteBudgetError.Message,
		)

		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteBudget ...
// @Summary Delete budget
// @Description Gets a list of all budgets current user is a part of
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Param Budget-ID header string true "Budget ID to delete"
// @Success 200 {array} models.Budget
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /budget/delete [delete]
func DeleteBudget(c *gin.Context) {

	budgetID, budgetIDError := uuid.Parse(c.GetHeader("Budget-ID"))

	if budgetIDError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Header 'Budget-ID' must contain a valid uuid",
		)

		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	deleteBudgetError := budgetService.DeleteBudget(budgetID, user.UserID)

	if deleteBudgetError != nil {
		requests.ThrowError(
			c,
			deleteBudgetError.StatusCode,
			deleteBudgetError.Error(),
		)

		return
	}

	c.Status(http.StatusNoContent)
}

// GetBudgetExpenseSummary ...
// @Summary Get Budget Expense summary
// @Description Gets data about user spending vs budget
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param Budget-ID header string true "Budget ID to get expense summary for"
// @Security Google AccessToken
// @Success 200 {object} string
// @Failure 403 {object} errors.Error
// @Router /budget/get-expense-summary [get]
func GetBudgetExpenseSummary(c *gin.Context) {
	budgetID, budgetIDError := uuid.Parse(c.GetHeader("Budget-ID"))

	if budgetIDError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Header 'Budget-ID' must contain a valid uuid",
		)

		return
	}

	user, getUserErr := requests.GetUserFromContext(c)

	if getUserErr != nil {
		panic(getUserErr)
	}

	summary, summaryErr := budgetService.GetBudgetExpenseSummary(budgetID, user.UserID)

	if summaryErr != nil {
		requests.ThrowError(
			c,
			summaryErr.StatusCode,
			summaryErr.Message,
		)

		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTransactionCategories ...
// @Summary Get Transaction Categories
// @Description Gets all transaction categories from plaid
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param Budget-ID header string true "Budget ID to get categories for"
// @Security Google AccessToken
// @Success 200 {array} models.BudgetTransactionCategory
// @Failure 403 {object} models.Error
// @Router /budget/transaction-categories [get]
func GetTransactionCategories(c *gin.Context) {
	budgetID, budgetIDError := uuid.Parse(c.GetHeader("Budget-ID"))

	if budgetIDError != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Header 'Budget-ID' must contain a valid uuid",
		)

		return
	}

	categories, err := budgetService.GetTransactionCategories(budgetID)

	if err != nil {
		requests.ThrowError(
			c,
			err.StatusCode,
			err.Message,
		)

		return
	}

	c.JSON(http.StatusOK, categories)
}

// DeleteBudgetTransactionCategory ...
// @Summary Delete budget transaction category
// @Description Deletes a budget transaction category
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Param budget-transaction-category-id path string true "Budget Transaction Category Id"
// @Success 200
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /budget/transaction-categories/delete/{budget-transaction-category-id} [delete]
func DeleteBudgetTransactionCategory(c *gin.Context) {
	param := c.Param("budget-transaction-category-id")
	budgetTransactionCategoryID, parseErr := uuid.Parse(param)

	if parseErr != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Budget Transaction Category ID must be a UUID",
		)

		return
	}

	error := budgetService.DeleteTransactionCategory(budgetTransactionCategoryID)

	if error != nil {
		requests.ThrowError(
			c,
			error.StatusCode,
			error.Message,
		)

		return
	}

	c.Status(http.StatusOK)
}

// CreateBudgetTransactionCategory ...
// @Summary Creates a budget transaction source
// @Description Creates a budget transaction source
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param budgetTransactionSource body models.BudgetTransactionCategoryCreationPayload true "Budget Transaction Category"
// @Security Google AccessToken
// @Success 200 {object} models.BudgetTransactionCategory
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /budget/transaction-categories/create [post]
func CreateBudgetTransactionCategory(c *gin.Context) {
	var json models.BudgetTransactionCategoryCreationPayload
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	category, creationErr := budgetService.CreateTransactionCategory(json)

	if creationErr != nil {
		requests.ThrowError(
			c,
			creationErr.StatusCode,
			creationErr.Message,
		)

		return
	}

	c.JSON(http.StatusOK, category)
}
