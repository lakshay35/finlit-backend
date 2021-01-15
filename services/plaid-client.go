package services

import (
	"fmt"
	"net/http"

	"github.com/plaid/plaid-go/plaid"
)

var (
	PLAID_CLIENT_ID     = "5f25d4dbd6e09e001026c0f6"
	PLAID_SECRET        = "690750c7b009518b14ffc6386b7fb4"
	PLAID_ENV           = "development"
	PLAID_COUNTRY_CODES = "US"
	PLAID_PRODUCTS      = "transactions"
	PLAID_REDIRECT_URI  = ""
	environments        = map[string]plaid.Environment{
		"sandbox":     plaid.Sandbox,
		"development": plaid.Development,
		"production":  plaid.Production,
	}
)

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
}()
