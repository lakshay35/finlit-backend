package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/services"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/lakshay35/finlit-backend/utils/requests"
	"github.com/plaid/plaid-go/plaid"
	uuid "github.com/satori/go.uuid"
)

// GetAccountInformation ...
// Gets account information based on access token
func GetAccountInformation(c *gin.Context) {

	var json Account
	err := c.BindJSON(&json)

	if err != nil {
		requests.ThrowError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := services.PlaidClient.GetAccounts(GetAccountAccessToken(json.ExternalAccountID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"error":   true,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully retrieved accounts",
		"data":    response.Accounts,
	})
}

// GetCurrentBalances ...
func GetCurrentBalances(c *gin.Context) {
	response, err := services.PlaidClient.GetBalances("accessToken")
	if err != nil {
		panic(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
	})
}

// GetTransactions ...
func GetTransactions(c *gin.Context) {
	// pull transactions for the past 30 days
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	response, err := services.PlaidClient.GetTransactions("accessToken", startDate, endDate)

	if err != nil {
		panic(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":     response.Accounts,
		"transactions": response.Transactions,
	})
}

// LinkTokenPayload ...
type LinkTokenPayload struct {
	LinkToken string `json:"linkToken"`
}

// CreateLinkToken ...
// Creates link token
func CreateLinkToken(c *gin.Context) {
	linkToken, err := linkTokenCreate(nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"error":   true,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successsfully created link token",
		"data": LinkTokenPayload{
			LinkToken: linkToken,
		},
	})
}

type httpError struct {
	errorCode int
	error     string
}

func (httpError *httpError) Error() string {
	return httpError.error
}

// linkTokenCreate creates a link token using the specified parameters
func linkTokenCreate(
	paymentInitiation *plaid.PaymentInitiation,
) (string, *httpError) {
	countryCodes := strings.Split(services.PLAID_COUNTRY_CODES, ",")
	products := strings.Split(services.PLAID_PRODUCTS, ",")
	redirectURI := services.PLAID_REDIRECT_URI
	configs := plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
			// This should correspond to a unique id for the current user.
			ClientUserID: "user-id",
		},
		ClientName:        "Plaid Quickstart",
		Products:          products,
		CountryCodes:      countryCodes,
		Language:          "en",
		RedirectUri:       redirectURI,
		PaymentInitiation: paymentInitiation,
	}
	resp, err := services.PlaidClient.CreateLinkToken(configs)
	if err != nil {
		return "", &httpError{
			errorCode: http.StatusBadRequest,
			error:     err.Error(),
		}
	}
	return resp.LinkToken, nil
}

// GetAccountAccessToken ...
// Get access token for an account based
// on accountID
func GetAccountAccessToken(accountID uuid.UUID) string {
	connection := database.GetConnection()

	query := "SELECT * FROM external_accounts WHERE external_account_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Error preparing statement while getting access token for accountID " + accountID.String())
	}

	var access_token string
	stmt.QueryRow(accountID).Scan(&access_token)

	return access_token
}

// Account ...
// Bank Account Integration Entity
type Account struct {
	ExternalAccountID uuid.UUID `json:"externalAccountID"`
	AccountName       string    `json:"accountName,omitempty"`
}

// GetAllAccounts ...
// Gets all accounts a user has
func GetAllAccounts(c *gin.Context) {
	connection := database.GetConnection()

	query := "SELECT * FROM external_accounts WHERE user_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		fmt.Println(err)
		panic("Error preparing statement to get user bank accounts")
	}

	user := requests.GetUserFromContext(c)

	rows, err := stmt.Query(user.UserID)

	if err != nil || !rows.NextResultSet() {
		requests.ThrowError(c, http.StatusNotFound, "User has no account registrations")
		return
	}

	var accounts []Account

	for rows.Next() {
		var external_account_id uuid.UUID
		var institutional_id string
		var user_id uuid.UUID
		var access_token string
		var account_name string

		rows.Scan(&access_token, external_account_id, &user_id, &account_name, &institutional_id)

		act := Account{
			ExternalAccountID: external_account_id,
			AccountName:       account_name,
		}

		accounts = append(accounts, act)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sucessfully retrieved accounts",
		"data":    accounts,
	})
}

type tokenPayload struct {
	Token string `json:"public_token"`
}

// RegisterAccessToken ...
// Exchanges public token for a
// permanent access token and stores
// it in the database
func RegisterAccessToken(c *gin.Context) {
	var json tokenPayload
	err := c.BindJSON(&json)

	if err != nil {
		panic("Unable to parse tokenpayload '/account/register'")
	}

	response, err := services.PlaidClient.ExchangePublicToken(json.Token)
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

	query := "INSERT INTO external_accounts (institutional_id, access_token, account_name, user_id) ($1, $2, $3, $4)"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Something went wrong while preparing query")
	}

	accounts := getAccountsForAccessToken(accessToken)

	for _, act := range accounts {
		_, err = stmt.Exec(act.AccountID, accessToken, act.Name, userObj.UserID)

		if err != nil {
			panic("Error occurred when looping over bank accounts from plaid")
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully registered permanent access token",
		"data":    nil,
	})

}

// getAccountsForAccessToken
// gets account tied to access token
// from Plaid APIS
func getAccountsForAccessToken(accessToken string) []plaid.Account {
	accounts, err := services.PlaidClient.GetAccounts(accessToken)

	if err != nil {
		panic("Something went wrong while getting accounts from plaid")
	}

	return accounts.Accounts
}
