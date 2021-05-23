package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	transactionService "github.com/lakshay35/finlit-backend/services/transaction"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// CategorizeExpense ...
// @Summary Categorizes a transaction to an expense category
// @Description Categorizes an transaction name to and expense category under the specified budget
// @Tags Transactions
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Param body body models.BudgetTransactionCategoryTransactionCreationPayload true "Payload"
// @Success 200 {array} models.BudgetTransactionCategoryTransactionCreationPayload
// @Failure 403 {object} models.Error
// @Router /transaction/categorize [post]
func CategorizeExpense(c *gin.Context) {
	var payload models.BudgetTransactionCategoryTransactionCreationPayload
	err := requests.ParseBody(c, &payload)

	if err != nil {
		return
	}

	categorizationErr := transactionService.CategorizeTransaction(payload)

	if categorizationErr != nil {
		requests.ThrowError(
			c,
			categorizationErr.StatusCode,
			categorizationErr.Message,
		)

		return
	}

	c.Status(http.StatusCreated)
}
