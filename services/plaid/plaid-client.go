package plaid

import (
	"fmt"
	"net/http"

	environment "github.com/lakshay35/finlit-backend/services/environment"
	"github.com/plaid/plaid-go/plaid"
)

var (
	PLAID_CLIENT_ID     = ""
	PLAID_SECRET        = ""
	PLAID_ENV           = ""
	PLAID_PRODUCTS      = ""
	PLAID_COUNTRY_CODES = ""
	PLAID_REDIRECT_URI  = ""
	environments        = map[string]plaid.Environment{
		"sandbox":     plaid.Sandbox,
		"development": plaid.Development,
		"production":  plaid.Production,
	}
)

func init() {
	PLAID_CLIENT_ID = environment.GetEnvVariable("PLAID_CLIENT_ID")
	PLAID_SECRET = environment.GetEnvVariable("PLAID_SECRET")
	PLAID_ENV = environment.GetEnvVariable("PLAID_ENV")
	PLAID_PRODUCTS = environment.GetEnvVariable("PLAID_PRODUCTS")
	PLAID_COUNTRY_CODES = environment.GetEnvVariable("PLAID_COUNTRY_CODES")
	PLAID_REDIRECT_URI = ""
}

var PlaidClient = func() *plaid.Client {
	client, err := plaid.NewClient(plaid.ClientOptions{
		PLAID_CLIENT_ID,
		PLAID_SECRET,
		environments[PLAID_ENV],
		&http.Client{},
	})
	if err != nil {
		panic(fmt.Errorf("unexpected error while initializing plaid client %w", err))
	}
	return client
}
