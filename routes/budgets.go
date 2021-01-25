package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	budgetService "github.com/lakshay35/finlit-backend/services/budget"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// CreateBudget ...
// @Summary Create a budget
// @Description Creates a budget with requesting user as owner
// @Tags Budgets
// @Accept  json
// @Produce  json
// @Param budget body models.Budget true "Budget body needed to create budget"
// @Security Google AccessToken
// @Success 200 {object} models.Expense
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 409 {object} models.Error
// @Router /budget/create [post]
func CreateBudget(c *gin.Context) {
	res := models.Budget{}
	err := budgetService.ParseBudget(c, &res)
	if err != nil || res.BudgetName == "" {
		requests.ThrowError(c, http.StatusBadRequest, "Error parsing body")
		return
	}

	user, errrr := c.Get("USER")

	if !errrr {
		panic("User object not found")
	}

	userObj := user.(models.User)

	if budgetService.DoesBudgetExist(userObj.UserID, res.BudgetName) {
		requests.ThrowError(c, http.StatusConflict, "Budget named "+res.BudgetName+" already exists")
		return
	}

	connection := database.GetConnection()

	query := "INSERT INTO budgets (owner_id, budget_name) VALUES ($1, $2)"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Something went wrong when preparing query to create budget")
	}

	_, errr = stmt.Exec(userObj.UserID, res.BudgetName)

	if errr != nil {
		panic(errr)
	}

	connection.Commit()

	c.JSON(201, budgetService.GetBudget(userObj.UserID, res.BudgetName))
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
// @Router /budget/get [post]
func GetBudgets(c *gin.Context) {
	connection := database.GetConnection()
	defer connection.Commit()

	user, err := c.Get("USER")

	if !err {
		panic("USER not found!")
	}

	userObj := user.(models.User)

	query := "SELECT * FROM budgets where owner_id = $1"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Error preparing query for getting budgets")
	}

	res, errr := stmt.Query(userObj.UserID)

	if errr != nil {
		requests.ThrowError(c, 404, "No budgets found")
		return
	}

	var result []models.Budget = make([]models.Budget, 0)

	for res.Next() {
		var temp models.Budget
		res.Scan(&temp.BudgetID, &temp.BudgetName, &temp.OwnerID)
		// Appends the item to the result
		result = append(result, temp)
	}

	c.JSON(http.StatusOK, result)
}
