package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	user := requests.GetUserFromContext(c)

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

	user := requests.GetUserFromContext(c)

	result, err := budgetService.GetAllBudgets(user.UserID)

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
