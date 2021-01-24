package models

// PlaidAccountBalances ...
// Plaid Account struct cause swaggo doesn't detect plaid.AccountBalances
type PlaidAccountBalances struct {
	Available              float64 `json:"available"`
	Current                float64 `json:"current"`
	Limit                  float64 `json:"limit"`
	ISOCurrencyCode        string  `json:"iso_currency_code"`
	UnofficialCurrencyCode string  `json:"unofficial_currency_code"`
}

// PlaidAccount ...
// Plaid Account struct cause swaggo doesn't detect plaid.Account
type PlaidAccount struct {
	AccountID          string               `json:"account_id"`
	Balances           PlaidAccountBalances `json:"balances"`
	Mask               string               `json:"mask"`
	Name               string               `json:"name"`
	OfficialName       string               `json:"official_name"`
	Subtype            string               `json:"subtype"`
	Type               string               `json:"type"`
	VerificationStatus string               `json:"verification_status"`
}

// PlaidGetBalancesResponse ...
// Plaid GetBalancesResponse struct cause swaggo doesn't detect plaid.GetBalancesResponse
type PlaidGetBalancesResponse struct {
	RequestID string         `json:"request_id"`
	Accounts  []PlaidAccount `json:"accounts"`
}

// PlaidTransaction ...
// Plaid Transaction struct cause swaggo doesn't detect plaid.GetBalancesResponse
type PlaidTransaction struct {
	AccountID              string   `json:"account_id"`
	Amount                 float64  `json:"amount"`
	ISOCurrencyCode        string   `json:"iso_currency_code"`
	UnofficialCurrencyCode string   `json:"unofficial_currency_code"`
	Category               []string `json:"category"`
	CategoryID             string   `json:"category_id"`
	Date                   string   `json:"date"`
	AuthorizedDate         string   `json:"authorized_date"`

	Location PlaidLocation `json:"location"`

	Name string `json:"name"`

	PaymentMeta    PlaidPaymentMeta `json:"payment_meta"`
	PaymentChannel string           `json:"payment_channel"`

	Pending              bool   `json:"pending"`
	PendingTransactionID string `json:"pending_transaction_id"`
	AccountOwner         string `json:"account_owner"`
	ID                   string `json:"transaction_id"`
	Type                 string `json:"transaction_type"`
	Code                 string `json:"transaction_code"`
}

// PlaidLocation ...
// Plaid Location struct cause swaggo doesn't detect plaid.GetBalancesResponse
type PlaidLocation struct {
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Region      string  `json:"region"`
	StoreNumber string  `json:"store_number"`
	PostalCode  string  `json:"postal_code"`
	Country     string  `json:"country"`
}

// PlaidPaymentMeta ...
// Plaid PaymentMeta struct cause swaggo doesn't detect plaid.GetBalancesResponse
type PlaidPaymentMeta struct {
	ByOrderOf        string `json:"by_order_of"`
	Payee            string `json:"payee"`
	Payer            string `json:"payer"`
	PaymentMethod    string `json:"payment_method"`
	PaymentProcessor string `json:"payment_processor"`
	PPDID            string `json:"ppd_id"`
	Reason           string `json:"reason"`
	ReferenceNumber  string `json:"reference_number"`
}
