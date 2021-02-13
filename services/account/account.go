package account

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lakshay35/finlit-backend/models"
	"github.com/lakshay35/finlit-backend/models/errors"
	plaidService "github.com/lakshay35/finlit-backend/services/plaid"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/plaid/plaid-go/plaid"
)

// GetAccountsForAccessToken ...
// gets account tied to access token
// from Plaid APIS
func GetAccountsForAccessToken(accessToken string) []plaid.Account {
	accounts, err := plaidService.PlaidClient().GetAccounts(accessToken)

	if err != nil {
		panic("Something went wrong while getting accounts from plaid")
	}

	return accounts.Accounts
}

// DeleteExternalAccount ...
// Deletes external account with provided id
func DeleteExternalAccount(accountID uuid.UUID) *errors.Error {

	_, getAccountErr := GetAccountInformation(accountID)

	if getAccountErr != nil {
		return getAccountErr
	}

	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "DELETE FROM external_accounts where external_account_id = $1"

	stmt := database.PrepareStatement(connection, query)

	_, deleteErr := stmt.Exec(accountID)

	if deleteErr != nil {
		return &errors.Error{
			Message:    deleteErr.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// GetExternalAccount ...
// Gets external account from DB
func GetExternalAccount(accountID uuid.UUID) (*models.Account, *errors.Error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT * FROM external_accounts WHERE external_account_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic(err)
	}

	var externalAccount models.Account

	err = stmt.QueryRow(accountID).Scan(
		&externalAccount.ExternalAccountID,
		&externalAccount.InstitutionalID,
		&externalAccount.UserID,
		&externalAccount.AccessToken,
		&externalAccount.AccountName,
	)

	if err != nil {
		return nil, &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return &externalAccount, nil
}

// LinkTokenCreate creates a link token using the specified parameters
func LinkTokenCreate(
	paymentInitiation *plaid.PaymentInitiation,
) (string, *errors.Error) {
	countryCodes := strings.Split(plaidService.PlaidCountryCodes, ",")
	products := strings.Split(plaidService.PlaidProducts, ",")
	redirectURI := plaidService.PlaidRedirectURI
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

	resp, err := plaidService.PlaidClient().CreateLinkToken(configs)

	if err != nil {
		return "", &errors.Error{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
	}

	return resp.LinkToken, nil
}

// GetAccountAccessToken ...
// Get access token for an account based
// on accountID
func GetAccountAccessToken(accountID uuid.UUID) string {
	connection := database.GetConnection()

	query := "SELECT access_token FROM external_accounts WHERE external_account_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Error preparing statement while getting access token for accountID " + accountID.String())
	}

	var access_token string

	err = stmt.QueryRow(accountID).Scan(&access_token)

	if err != nil {
		panic(err)
	}

	return access_token
}

// RegisterExternalAccounts ...
// Registers external accounts in the database based on acces token
func RegisterExternalAccounts(accessToken string, userID uuid.UUID) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "INSERT INTO external_accounts (institutional_id, access_token, account_name, user_id) VALUES ($1, $2, $3, $4)"

	stmt, err := connection.Prepare(query)

	if err != nil {
		fmt.Println(err)
		panic("Something went wrong while preparing query")
	}

	accounts := GetAccountsForAccessToken(accessToken)

	for _, act := range accounts {
		_, err = stmt.Exec(act.AccountID, accessToken, act.OfficialName+" "+act.Name, userID)

		if err != nil {
			panic("Error occurred when looping over bank accounts from plaid")
		}
	}
}

// GetAllExternalAccounts ...
// Gets all external accounts tagged with the given userID
func GetAllExternalAccounts(userID uuid.UUID) ([]models.Account, *errors.Error) {
	connection := database.GetConnection()

	defer database.CloseConnection(connection)

	query := "SELECT * FROM external_accounts WHERE user_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		fmt.Println(err)
		panic("Error preparing statement to get user bank accounts")
	}

	rows, err := stmt.Query(userID)

	if err != nil {
		return nil, &errors.Error{
			Message:    "User has no account registrations",
			StatusCode: http.StatusNotFound,
		}
	}

	var accounts []models.Account = make([]models.Account, 0)

	for rows.Next() {
		var temp models.Account

		err = rows.Scan(
			&temp.ExternalAccountID,
			&temp.InstitutionalID,
			&temp.UserID,
			&temp.AccessToken,
			&temp.AccountName,
		)

		if err != nil {
			panic(err)
		}

		accounts = append(accounts, temp)
	}

	rows.Close()

	return accounts, nil
}

// RegisterAccessToken ...
// Registers access token after exchanging public token for given userID
func RegisterAccessToken(token string, userID uuid.UUID) *errors.Error {
	response, err := plaidService.PlaidClient().ExchangePublicToken(token)
	if err != nil {
		return &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	accessToken := response.AccessToken

	RegisterExternalAccounts(accessToken, userID)

	return nil
}

// GetTransactions ...
// Gets Transactions for specified time period
func GetTransactions(
	externalAccountID uuid.UUID,
	startDate string,
	endDate string,
) ([]plaid.Transaction, *errors.Error) {

	externalAccount, GetExternalAccountErr := GetExternalAccount(externalAccountID)

	if GetExternalAccountErr != nil {
		return nil, &errors.Error{
			Message:    "External Acount not found",
			StatusCode: http.StatusBadRequest,
		}
	}

	response, err := plaidService.PlaidClient().GetTransactions(externalAccount.AccessToken, startDate, endDate)

	if err != nil {
		return nil, &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	var transactions []plaid.Transaction

	for _, tx := range response.Transactions {
		if strings.EqualFold(tx.AccountID, externalAccount.InstitutionalID) {
			transactions = append(transactions, tx)
		}
	}

	return transactions, nil
}

// GetAccountInformation ...
// Gets account information based on external account ID
func GetAccountInformation(
	externalAccountID uuid.UUID,
) (*plaid.Account, *errors.Error) {
	externalAccount, GetExternalAccountErr := GetExternalAccount(externalAccountID)

	if GetExternalAccountErr != nil {
		return nil, &errors.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "External Account not found",
		}
	}

	response, err := plaidService.PlaidClient().GetAccounts(externalAccount.AccessToken)

	if err != nil {
		return nil, &errors.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	var account plaid.Account

	for _, act := range response.Accounts {
		if strings.EqualFold(act.AccountID, externalAccount.InstitutionalID) {
			account = act
		}
	}

	fmt.Println(account)

	return &account, nil
}
