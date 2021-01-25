package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	roleService "github.com/lakshay35/finlit-backend/services/role"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

type addRoleRequest struct {
	Role     string    `json:"role"`
	UserID   uuid.UUID `json:"userID"`
	BudgetID uuid.UUID `json:"budgetId"`
}

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

	user, found := c.Get("USER")

	userObj := user.(models.User)

	res := addRoleRequest{}
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		// TODO: Create a request body interceptor annotation
		requests.ThrowError(c, 400, "malformed body struct")
		return
	}

	json.Unmarshal(jsonData, &res)

	if !found {
		panic("User ID not found in request context")
	}

	// Only budget owner can add users to budget
	if !roleService.DoesUserOwnBudget(userObj.UserID, res.BudgetID) {
		requests.ThrowError(c, 401, "Not enough permissions to add users to")
	}

	connection := database.GetConnection()
	defer connection.Commit()

	query := "INSERT INTO user_roles (user_id, role_id, budget_id) VALUES ('$1, (select role_id from roles where role_name = $2), $3)"

	stmt := database.PrepareStatement(connection, query)

	_, err = stmt.Exec(res.UserID, roleService.GetRole(res.Role), res.BudgetID)

	if err != nil {
		requests.ThrowError(c, 400, "Unknw")
	}

	c.Status(http.StatusOK)
}
