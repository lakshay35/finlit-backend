package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	accountsService "github.com/lakshay35/finlit-backend/services/account"
	plaidService "github.com/lakshay35/finlit-backend/services/plaid"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
	"github.com/plaid/plaid-go/plaid"
)

// GetAccountInformation ...
// @Summary Get Get Account Information
// @Description Gets account information based on access token
// @Tags account
// @Accept  json
// @Param account body Account true "Account payload to get informaion on"
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} models.PlaidAccount
// @Failure  400 {object} models.Error
// @Router /account/get-account-details [get]
func GetAccountInformation(c *gin.Context) {

	var json Account
	err := c.BindJSON(&json)

	if err != nil {
		requests.ThrowError(c, http.StatusBadRequest, err.Error())
		return
	}

	institutionalID, accessToken := accountsService.GetExternalAccount(json.ExternalAccountID)

	response, err := plaidService.PlaidClient().GetAccounts(accessToken)

	if err != nil {
		panic(err)
	}

	var account plaid.Account

	for _, act := range response.Accounts {
		if strings.EqualFold(act.AccountID, institutionalID) {
			account = act
		}
	}

	c.JSON(http.StatusOK, account)
}

// GetCurrentBalances ...
// @Summary Get Current A/c Balances
// @Description Retrieves live account balances for all accounts attached to an external account registratiom
// @Tags account
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} models.PlaidGetBalancesResponse
// @Failure  400 {object} models.Error
// @Router /account/live-balances [get]
func GetCurrentBalances(c *gin.Context) {
	response, err := plaidService.PlaidClient().GetBalances("accessToken")
	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
	}

	c.JSON(http.StatusOK, response.Accounts)
}

// GetTransactions ...
// @Summary Get Transactions
// @Description Gets all transactions for the  past 30 days
// @Tags account
// @Accept  json
// @Produce  json
// @Param body body Account true "Account payload to identify transactions with"
// @Security ApiKeyAuth
// @Success 200 {array} models.PlaidTransaction
// @Failure 403 {object} models.Error
// @Router /account/transactions [get]
func GetTransactions(c *gin.Context) {
	var json Account
	c.BindJSON(&json)

	// pull transactions for the past 30 days
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	institutionalID, accessToken := accountsService.GetExternalAccount(json.ExternalAccountID)

	response, err := plaidService.PlaidClient().GetTransactions(accessToken, startDate, endDate)

	if err != nil {
		panic(err)
	}

	var transactions []plaid.Transaction

	for _, tx := range response.Transactions {
		if strings.EqualFold(tx.AccountID, institutionalID) {
			transactions = append(transactions, tx)
		}
	}

	c.JSON(http.StatusOK, transactions)
}

// LinkTokenPayload ...
type LinkTokenPayload struct {
	LinkToken string `json:"linkToken"`
}

// CreateLinkToken ...
// Creates link token
// @Summary Create Link Token
// @Description Creates a link token to setup UI for generating public tokens
// @Tags account
// @Accept  json
// @Param id path string true "Expense ID (UUID)"
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} LinkTokenPayload
// @Failure 400 {object} models.Error
// @Router /budget/create [post]
func CreateLinkToken(c *gin.Context) {
	linkToken, err := accountsService.LinkTokenCreate(nil)
	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	c.JSON(http.StatusOK, LinkTokenPayload{
		LinkToken: linkToken,
	})
}

type httpError struct {
	errorCode int
	error     string
}

func (httpError *httpError) Error() string {
	return httpError.error
}

// Account ...
// Bank Account Integration Entity
type Account struct {
	ExternalAccountID uuid.UUID `json:"external_account_id"`
	AccountName       string    `json:"account_name,omitempty"`
}

// GetAllAccounts ...
// @Summary Get all registered external accounts
// @Description Gets a list of all external accounts registered via Plaid
// @Tags account
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} Account
// @Failure 403 {object} models.Error
// @Failure 404 {object} models.Error
// @Router /account/get [get]
func GetAllAccounts(c *gin.Context) {
	connection := database.GetConnection()

	query := "SELECT account_name, external_account_id FROM external_accounts WHERE user_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		fmt.Println(err)
		panic("Error preparing statement to get user bank accounts")
	}

	user := requests.GetUserFromContext(c)

	rows, err := stmt.Query(user.UserID)

	if err != nil {
		requests.ThrowError(c, http.StatusNotFound, "User has no account registrations")
		return
	}

	var accounts []Account = make([]Account, 0)

	for rows.Next() {
		var temp Account

		rows.Scan(&temp.AccountName, &temp.ExternalAccountID)

		if err != nil {
			panic(err)
		}

		accounts = append(accounts, temp)
	}

	c.JSON(http.StatusOK, accounts)
}

type tokenPayload struct {
	Token string `json:"public_token"`
}

// RegisterAccessToken ...
// @Summary Register Access Token
// @Description Creates a permanent access token based on public token
// @Tags account
// @Accept  json
// @Param body body tokenPayload true "Token Payload for registering access token"
// @Produce  json
// @Security ApiKeyAuth
// @Success 201
// @Failure 403 {object} models.Error
// @Router /account/register-token [post]
func RegisterAccessToken(c *gin.Context) {
	var json tokenPayload
	err := c.BindJSON(&json)

	if err != nil {
		panic("Unable to parse tokenpayload '/account/register'")
	}

	response, err := plaidService.PlaidClient().ExchangePublicToken(json.Token)
	if err != nil {
		requests.ThrowError(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken := response.AccessToken

	user, exists := c.Get("USER")

	if !exists {
		panic("USER NOT FOUND IN CONTEXT")
	}

	userObj := user.(models.User)

	connection := database.GetConnection()

	defer connection.Commit()

	query := "INSERT INTO external_accounts (institutional_id, access_token, account_name, user_id) VALUES ($1, $2, $3, $4)"

	stmt, err := connection.Prepare(query)

	if err != nil {
		fmt.Println(err)
		panic("Something went wrong while preparing query")
	}

	accounts := accountsService.GetAccountsForAccessToken(accessToken)

	for _, act := range accounts {
		_, err = stmt.Exec(act.AccountID, accessToken, act.OfficialName+" "+act.Name, userObj.UserID)

		if err != nil {
			panic("Error occurred when looping over bank accounts from plaid")
		}
	}

	c.JSON(http.StatusCreated, nil)

}
