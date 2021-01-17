package services

import (
	"fmt"
	"net/http"
	"os"

	"github.com/plaid/plaid-go/plaid"
)

var (
	PLAID_CLIENT_ID     = os.Getenv("PLAID_CLIENT_ID")
	PLAID_SECRET        = os.Getenv("PLAID_SECRET")
	PLAID_ENV           = os.Getenv("PLAID_ENV")
	PLAID_PRODUCTS      = os.Getenv("PLAID_PRODUCTS")
	PLAID_COUNTRY_CODES = os.Getenv("PLAID_COUNTRY_CODES")
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
