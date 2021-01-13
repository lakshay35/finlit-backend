package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type budget struct {
	BudgetName string `json:"budgetName"`
}

type budgetResponse struct {
	BudgetName string    `json:"budgetName"`
	OwnerID    uuid.UUID `json:"ownerID"`
	BudgetID   uuid.UUID `json:"budgetID"`
}

// CreateBudget ...
// Create a Budget
func CreateBudget(c *gin.Context) {
	res := budget{}
	err := parseBudget(c, &res)
	if err != nil {
		return
	}

	user, errrr := c.Get("USER")

	if !errrr {
		panic("User object not found")
	}

	userObj := user.(User)

	if doesBudgetExist(userObj.UserID, res.BudgetName) {
		ThrowError(c, http.StatusConflict, "Budget named "+res.BudgetName+" already exists")
		return
	}

	connection := GetConnection()

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

	c.JSON(201, gin.H{
		"message": "Successfully Created Budget",
		"data":    getBudget(userObj.UserID, res.BudgetName),
	})
}

// GetBudgets ...
// gets budgets pertaining
// to current user
func GetBudgets(c *gin.Context) {
	connection := GetConnection()
	defer connection.Commit()

	user, err := c.Get("USER")

	if !err {
		panic("USER not found!")
	}

	userObj := user.(User)

	query := "SELECT * FROM budgets where owner_id = $1"

	stmt, errr := connection.Prepare(query)

	if errr != nil {
		panic("Error preparing query for getting budgets")
	}

	res, errr := stmt.Query(userObj.UserID)

	if errr != nil {
		ThrowError(c, 404, "No budgets found")
		return
	}

	var result []budgetResponse

	var budget_name string
	var budget_id uuid.UUID
	var owner_id uuid.UUID

	for res.Next() {
		var temp budgetResponse
		res.Scan(&budget_id, &budget_name, &owner_id)
		temp.BudgetID = budget_id
		temp.BudgetName = budget_name
		temp.OwnerID = owner_id
		// Appends the item to the result
		result = append(result, temp)
	}

	c.JSON(200, gin.H{
		"message": "Successfully got all your budgets",
		"data":    result,
	})
}

// getBudget ...
// Gets budget from db based
// on given params
func getBudget(userID uuid.UUID, budgetName string) budgetResponse {
	connection := GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner_id = $1 AND budget_name = $2"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Something went wrong when preparing query to get budget")
	}

	rows, err := stmt.Query(userID, budgetName)

	if err != nil || !rows.Next() {
		fmt.Println(err.Error())
		return budgetResponse{}
	}

	var res budgetResponse
	var budget_id uuid.UUID
	var owner_id uuid.UUID
	var budget_name string

	rows.Scan(&budget_id, &owner_id, &budget_name)

	res.BudgetID = budget_id
	res.OwnerID = userID
	res.BudgetName = budgetName

	return res
}

// doesBudgetExist ...
// Checks if a budget exists
func doesBudgetExist(UserID uuid.UUID, budgetName string) bool {
	connection := GetConnection()
	defer connection.Commit()

	query := "SELECT * FROM budgets WHERE owner_id = $1 AND budget_name = $2"

	stmt, err := connection.Prepare(query)

	if err != nil {
		return false
	}

	fmt.Println("Queried db for dudget named " + budgetName)
	res, err := stmt.Query(UserID, budgetName)

	if err != nil {
		panic(err.Error())
	}
	return res.Next()
}

// parseBudget ...
// Parses body to budget{} type
// Throws error if body does not match
func parseBudget(c *gin.Context, res *budget) error {
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonData, &res)

	if err != nil {
		ThrowError(c, 400, "request body structure match error")
		return err
	}

	return nil
}
