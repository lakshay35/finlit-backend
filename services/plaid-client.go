package services

import (
	"fmt"
	"net/http"

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
	PLAID_CLIENT_ID = GetEnvVariable("PLAID_CLIENT_ID")
	PLAID_SECRET = GetEnvVariable("PLAID_SECRET")
	PLAID_ENV = GetEnvVariable("PLAID_ENV")
	PLAID_PRODUCTS = GetEnvVariable("PLAID_PRODUCTS")
	PLAID_COUNTRY_CODES = GetEnvVariable("PLAID_COUNTRY_CODES")
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
