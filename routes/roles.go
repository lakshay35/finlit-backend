package routes

import (
	"net/http"

	"github.com/lakshay35/finlit-backend/models"

	"github.com/gin-gonic/gin"
	roleService "github.com/lakshay35/finlit-backend/services/role"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// AddUserRoleToBudget ...
// @Summary Registers user to the database
// @Description Registers a user profile in the finlit database
// @Tags User Roles
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200
// @Failure 403 {object} models.Error
// @Failure 409 {object} models.Error
// @Router /role/add-user-role-to-budget [post]
func AddUserRoleToBudget(c *gin.Context) {

	var json models.AddRolePayload
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	user := requests.GetUserFromContext(c)

	roleService.AddRoleToBudget(user.UserID, json.BudgetID, json.Role)

	c.Status(http.StatusOK)
}
