package expense

import (
	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/utils/database"
)

// GetExpense ...
// Gets expense based on expense_id
func GetExpense(id uuid.UUID) (*models.Expense, error) {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT * FROM expenses WHERE expense_id = $1"

	stmt := database.PrepareStatement(connection, query)

	var expense models.Expense

	var expense_charge_cycle_id int

	err := stmt.QueryRow(id).Scan(
		&expense.ExpenseID,
		&expense.BudgetID,
		&expense.ExpenseName,
		&expense.ExpenseValue,
		&expense.ExpenseDescription,
		&expense_charge_cycle_id,
	)

	unit, err := GetExpenseChargeCycleName(expense_charge_cycle_id)

	if err != nil {
		panic(err)
	}

	expense.ExpenseChargeCycle = unit

	if err != nil {
		return &models.Expense{}, err
	}

	return &expense, nil
}

// IsUserAdmin ...
// Determines if user is an admin on the given budget
func IsUserAdmin(budgetID uuid.UUID, userID uuid.UUID) bool {
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

// IsUserViewer ...
// Determines if user is an admin on the given budget
func IsUserViewer(budgetID uuid.UUID, userID uuid.UUID) bool {
	connection := database.GetConnection()

	defer connection.Commit()

	query := `SELECT * FROM user_roles ur WHERE ur.user_id = $1 AND ur.role_id = (SELECT role_id FROM roles WHERE role_name = 'View Rights')
	AND ur.budget_id = $2`

	stmt := database.PrepareStatement(connection, query)

	rows, err := stmt.Query(userID, budgetID)

	if err != nil {
		panic(err)
	}

	return rows.Next()
}

// IsUserOwner ...
// Checks if user is an owner
func IsUserOwner(budgetID uuid.UUID, userID uuid.UUID) bool {
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

// GetExpenseChargeCycleID ...
// Gets expense charge cycle id for a given expense name
func GetExpenseChargeCycleID(expenseName string) (int, error) {
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

// GetExpenseChargeCycleName ...
// Gets the name of an expense charge cycle based on id
func GetExpenseChargeCycleName(expenseChargeCycleID int) (string, error) {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT unit from expense_charge_cycles WHERE expense_charge_cycle_id = $1"

	stmt := database.PrepareStatement(connection, query)

	var unit string

	err := stmt.QueryRow(expenseChargeCycleID).Scan(&unit)

	if err != nil {
		return "", err
	}

	return unit, nil
}
