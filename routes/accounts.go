package routes

import (
	"net/http"
	"time"

	uuid "github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"

	"github.com/gin-gonic/gin"
	accountsService "github.com/lakshay35/finlit-backend/services/account"
	plaidService "github.com/lakshay35/finlit-backend/services/plaid"
	"github.com/lakshay35/finlit-backend/utils/logging"
	"github.com/lakshay35/finlit-backend/utils/requests"
)

// GetAccountInformation ...
// @Summary Get Account Information
// @Description Gets account information based on access token
// @Tags External Accounts
// @Accept  json
// @Param account body models.Account true "Account payload to get informaion on"
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.PlaidAccount
// @Failure  400 {object} models.Error
// @Router /account/get-account-details [post]
func GetAccountInformation(c *gin.Context) {

	var json models.Account
	err := requests.ParseBody(c, &json)

	if err != nil {
		return
	}

	account, getAccountInformationError := accountsService.GetAccountInformation(json.ExternalAccountID)

	if getAccountInformationError != nil {
		requests.ThrowError(
			c,
			getAccountInformationError.StatusCode,
			getAccountInformationError.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, account)
}

// GetCurrentBalances ...
// @Summary Get Current A/c Balances
// @Description Retrieves live account balances for all accounts attached to an external account registratiom
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Security Google AccessToken
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

// DeleteAccount ...
// @Summary Delete Account
// @Description Deletes an external account registration
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Param external-account-id path string true "External Account Id"
// @Security Google AccessToken
// @Success 200 {array} models.Budget
// @Failure 403 {object} models.Error
// @Failure 400 {object} models.Error
// @Router /account/delete/{external-account-id} [delete]
func DeleteAccount(c *gin.Context) {

	param := c.Param("external-account-id")

	logging.InfoLogger.Print("Received request to delete external account with id ", param)
	externalAccountID, err := uuid.Parse(param)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

		return
	}

	deleteErr := accountsService.DeleteExternalAccount(externalAccountID)

	if deleteErr != nil {
		requests.ThrowError(
			c,
			deleteErr.StatusCode,
			deleteErr.Message,
		)

		return
	}

	c.Status(http.StatusOK)
}

// GetAccountByID ...
// @Summary Get Account by id
// @Description Deletes a budget transaction source
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Param external-account-id path string true "External Account Id"
// @Security Google AccessToken
// @Success 200 {object} models.Account
// @Failure 404 {object} models.Error
// @Failure 400 {object} models.Error
// @Failure 403 {object} models.Error
// @Router /account/get-account/{external-account-id} [get]
func GetAccountByID(c *gin.Context) {
	param := c.Param("external-account-id path")
	externalAccountID, parseErr := uuid.Parse(param)

	if parseErr != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			"Budget Transaction Source ID must be a UUID",
		)

		return
	}

	account, getErr := accountsService.GetExternalAccount(externalAccountID)

	if getErr != nil {
		requests.ThrowError(
			c,
			getErr.StatusCode,
			getErr.Message,
		)

		return
	}

	c.JSON(http.StatusOK, account)
}

// GetTransactions ...
// @Summary Get Transactions
// @Description Gets all transactions for the  past 30 days
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Param body body models.Account true "Account payload to identify transactions with"
// @Security Google AccessToken
// @Success 200 {array} models.PlaidTransaction
// @Failure 403 {object} models.Error
// @Router /account/transactions [post]
func GetTransactions(c *gin.Context) {
	var json models.Account
	err := requests.ParseBody(
		c,
		&json,
	)

	if err != nil {
		return
	}

	transactions, transactionsError := accountsService.GetTransactions(
		json.ExternalAccountID,
		time.Now().Local().Add(-30*24*time.Hour).Format("2006-01-02"),
		time.Now().Local().Format("2006-01-02"),
	)

	if transactionsError != nil {
		requests.ThrowError(
			c,
			transactionsError.StatusCode,
			transactionsError.Message,
		)

		return
	}

	c.JSON(
		http.StatusOK,
		transactions,
	)
}

// RenewAccessToken ...
// Renews Access token
// @Summary Renew Access Token
// @Description Creates a link token to setup UI for renewing access tokens
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Param account body models.AccessTokenPayload true "Access Token payload"
// @Success 200 {object} models.AccountIdPayload
// @Failure 400 {object} models.Error
// @Router /account/create-link-token [get]
func RenewAccessToken(c *gin.Context) {
	var body models.AccountIdPayload

	parseError := requests.ParseBody(c, &body)

	if parseError != nil {
		return
	}

	user, userError := requests.GetUserFromContext(c)

	if userError != nil {
		panic(userError)
	}

	c.JSON(http.StatusOK, &models.LinkTokenPayload{
		LinkToken: accountsService.GetAccessTokenRenewalLinkToken(user.UserID.String(), body.ExternalAccountID),
	})
}

// CreateLinkToken ...
// Creates link token
// @Summary Create Link Token
// @Description Creates a link token to setup UI for generating public tokens
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {object} models.LinkTokenPayload
// @Failure 400 {object} models.Error
// @Router /account/create-link-token [get]
func CreateLinkToken(c *gin.Context) {
	user, userError := requests.GetUserFromContext(c)

	if userError != nil {
		panic(userError)
	}

	linkToken, err := accountsService.LinkTokenCreate(nil, user.GoogleID)

	if err != nil {
		requests.ThrowError(
			c,
			err.StatusCode,
			err.Error(),
		)
		return
	}

	c.JSON(http.StatusOK, models.LinkTokenPayload{
		LinkToken: linkToken,
	})
}

// GetAllAccounts ...
// @Summary Get all registered external accounts
// @Description Gets a list of all external accounts registered via Plaid
// @Tags External Accounts
// @Accept  json
// @Produce  json
// @Security Google AccessToken
// @Success 200 {array} models.Account
// @Failure 403 {object} models.Error
// @Failure 404 {object} models.Error
// @Router /account/get [get]
func GetAllAccounts(c *gin.Context) {
	user, err := requests.GetUserFromContext(c)

	if err != nil {
		panic(err)
	}

	accounts, err := accountsService.GetAllExternalAccounts(user.UserID)

	if err != nil {
		requests.ThrowError(
			c,
			err.StatusCode,
			err.Error(),
		)
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// RegisterAccessToken ...
// @Summary Register Access Token
// @Description Creates a permanent access token based on public token
// @Tags External Accounts
// @Accept  json
// @Param body body models.PublicTokenPayload true "Token Payload for registering access token"
// @Produce  json
// @Security Google AccessToken
// @Success 201
// @Failure 403 {object} models.Error
// @Router /account/register-token [post]
func RegisterAccessToken(c *gin.Context) {
	var json models.PublicTokenPayload
	err := c.BindJSON(&json)

	if err != nil {
		requests.ThrowError(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	user, err := requests.GetUserFromContext(c)

	if err != nil {
		panic(err)
	}

	accessTokenError := accountsService.RegisterAccessToken(json.Token, user.UserID)

	if accessTokenError != nil {
		requests.ThrowError(
			c,
			accessTokenError.StatusCode,
			accessTokenError.Error(),
		)
		return
	}

	c.JSON(http.StatusCreated, nil)
}
