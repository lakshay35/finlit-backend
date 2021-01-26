package plaid

import (
	"fmt"
	"net/http"

	environment "github.com/lakshay35/finlit-backend/services/environment"
	"github.com/plaid/plaid-go/plaid"
)

// PlaidClientID
var (
	PlaidClientID     = ""
	PlaidSecret       = ""
	PlaidEnv          = ""
	PlaidProducts     = ""
	PlaidCountryCodes = ""
	PlaidRedirectURI  = ""
	environments      = map[string]plaid.Environment{
		"sandbox":     plaid.Sandbox,
		"development": plaid.Development,
		"production":  plaid.Production,
	}
)

func init() {
	PlaidClientID = environment.GetEnvVariable("PLAID_CLIENT_ID")
	PlaidSecret = environment.GetEnvVariable("PLAID_SECRET")
	PlaidEnv = environment.GetEnvVariable("PLAID_ENV")
	PlaidProducts = environment.GetEnvVariable("PLAID_PRODUCTS")
	PlaidCountryCodes = environment.GetEnvVariable("PLAID_COUNTRY_CODES")
	PlaidRedirectURI = ""
}

// PlaidClient ...
// Client to communicate with plaid api
var PlaidClient = func() *plaid.Client {
	client, err := plaid.NewClient(
		plaid.ClientOptions{
			ClientID:    PlaidClientID,
			Secret:      PlaidSecret,
			Environment: environments[PlaidEnv],
			HTTPClient:  &http.Client{},
		})
	if err != nil {
		panic(fmt.Errorf("unexpected error while initializing plaid client %w", err))
	}
	return client
}
