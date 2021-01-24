package account

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	plaidService "github.com/lakshay35/finlit-backend/services/plaid"
	"github.com/lakshay35/finlit-backend/utils/database"
	"github.com/plaid/plaid-go/plaid"
)

type httpError struct {
	errorCode int
	error     string
}

func (httpError *httpError) Error() string {
	return httpError.error
}

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

// GetExternalAccount ...
// Gets external account from DB
func GetExternalAccount(accountID uuid.UUID) (string, string) {
	connection := database.GetConnection()

	defer connection.Commit()

	query := "SELECT institutional_id, access_token FROM external_accounts WHERE external_account_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic(err)
	}

	var institutional_id string
	var access_token string

	stmt.QueryRow(accountID).Scan(&institutional_id, &access_token)

	return institutional_id, access_token
}

// LinkTokenCreate creates a link token using the specified parameters
func LinkTokenCreate(
	paymentInitiation *plaid.PaymentInitiation,
) (string, *httpError) {
	countryCodes := strings.Split(plaidService.PLAID_COUNTRY_CODES, ",")
	products := strings.Split(plaidService.PLAID_PRODUCTS, ",")
	redirectURI := plaidService.PLAID_REDIRECT_URI
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

	query := "SELECT access_token FROM external_accounts WHERE external_account_id = $1"

	stmt, err := connection.Prepare(query)

	if err != nil {
		panic("Error preparing statement while getting access token for accountID " + accountID.String())
	}

	var access_token string

	stmt.QueryRow(accountID).Scan(&access_token)

	return access_token
}
