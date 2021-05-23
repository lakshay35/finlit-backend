package transaction

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	budgetService "github.com/lakshay35/finlit-backend/services/budget"
	"github.com/lakshay35/finlit-backend/utils/database"
)

// CategorizeTransaction ...
func CategorizeTransaction(payload models.BudgetTransactionCategoryTransactionCreationPayload) *errors.Error {
	transactionCategories, transactionCategoriesErr := budgetService.GetTransactionCategories(payload.BudgetID)

	if transactionCategoriesErr != nil {
		return transactionCategoriesErr
	}

	var categoryID uuid.UUID
	categoryValid := false

	for _, cat := range transactionCategories {
		if strings.EqualFold(cat.CategoryName, payload.CategoryName) {
			categoryID = cat.BudgetTransactionCategoryID
			categoryValid = true
		}
	}

	if !categoryValid {
		return &errors.Error{
			Message:    "Provided category not found",
			StatusCode: http.StatusBadRequest,
		}
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "INSERT INTO budget_transaction_category_transactions (budget_transaction_category_id, transaction_name) VALUES ($1, $2)"

	stmt := database.PrepareStatement(connection, query)

	_, insertErr := stmt.Exec(categoryID, payload.TransactionName)

	if insertErr != nil {
		return &errors.Error{
			Message:    insertErr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}
