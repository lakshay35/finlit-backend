package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/plaid/plaid-go/plaid"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"google.golang.org/api/oauth2/v2"
)

func init() {
	if PLAID_PRODUCTS == "" {
		PLAID_PRODUCTS = "transactions"
	}
	if PLAID_COUNTRY_CODES == "" {
		PLAID_COUNTRY_CODES = "US"
	}
	if PLAID_ENV == "" {
		PLAID_ENV = "sandbox"
	}
}

// Fill with your Plaid API keys - https://dashboard.plaid.com/account/keys
var (
	PLAID_CLIENT_ID     = "5f25d4dbd6e09e001026c0f6"
	PLAID_SECRET        = "690750c7b009518b14ffc6386b7fb4"
	PLAID_ENV           = "development"
	PLAID_PRODUCTS      = "auth,transactions"
	PLAID_COUNTRY_CODES = os.Getenv("PLAID_COUNTRY_CODES")
	// Parameters used for the OAuth redirect Link flow.
	//
	// Set PLAID_REDIRECT_URI to 'http://localhost:8000/oauth-response.html'
	// The OAuth redirect flow requires an endpoint on the developer's website
	// that the bank website should redirect to. You will need to configure
	// this redirect URI for your client ID through the Plaid developer dashboard
	// at https://dashboard.plaid.com/team.
	PLAID_REDIRECT_URI = "https://localhost:3000"

	// Use 'sandbox' to test with fake credentials in Plaid's Sandbox environment
	// Use `development` to test with real credentials while developing
	// Use `production` to go live with real users
	APP_PORT = os.Getenv("PORT")
)

var environments = map[string]plaid.Environment{
	"sandbox":     plaid.Sandbox,
	"development": plaid.Development,
	"production":  plaid.Production,
}

var client = func() *plaid.Client {
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

// We store the access_token in memory - in production, store it in a secure
// persistent data store.
var accessToken string = "access-development-1e936067-accd-4af9-a21c-73fcf2b8f1fd"
var itemID string

var paymentID string

func renderError(c *gin.Context, err error) {
	if plaidError, ok := err.(plaid.Error); ok {
		// Return 200 and allow the front end to render the error.
		c.JSON(http.StatusOK, gin.H{"error": plaidError})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func getAccessToken(c *gin.Context) {
	publicToken := c.PostForm("public_token")
	response, err := client.ExchangePublicToken(publicToken)
	if err != nil {
		renderError(c, err)
		return
	}
	accessToken = response.AccessToken
	itemID = response.ItemID

	fmt.Println("public token: " + publicToken)
	fmt.Println("access token: " + accessToken)
	fmt.Println("item ID: " + itemID)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"item_id":      itemID,
	})
}

// This functionality is only relevant for the UK Payment Initiation product.
// Creates a link token configured for payment initiation. The payment
// information will be associated with the link token, and will not have to be
// passed in again when we initialize Plaid Link.
func createLinkTokenForPayment(c *gin.Context) {
	recipientCreateResp, err := client.CreatePaymentRecipient(
		"Harry Potter",
		"GB33BUKB20201555555555",
		&plaid.PaymentRecipientAddress{
			Street:     []string{"4 Privet Drive"},
			City:       "Little Whinging",
			PostalCode: "11111",
			Country:    "GB",
		})
	if err != nil {
		renderError(c, err)
		return
	}
	paymentCreateResp, err := client.CreatePayment(recipientCreateResp.RecipientID, "payment-ref", plaid.PaymentAmount{
		Currency: "GBP",
		Value:    12.34,
	})
	if err != nil {
		renderError(c, err)
		return
	}
	paymentID = paymentCreateResp.PaymentID
	fmt.Println("payment id: " + paymentID)

	linkToken, tokenCreateErr := linkTokenCreate(&plaid.PaymentInitiation{
		PaymentID: paymentID,
	})
	if tokenCreateErr != nil {
		renderError(c, tokenCreateErr)
	}
	c.JSON(http.StatusOK, gin.H{
		"link_token": linkToken,
	})
}

func auth(c *gin.Context) {
	response, err := client.GetAuth(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
		"numbers":  response.Numbers,
	})
}

// @Description Summary
func accounts(c *gin.Context) {
	response, err := client.GetAccounts(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
	})
}

func getCategories(c *gin.Context) {
	response, err := client.GetCategories()

	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": response.Categories,
	})
}

func balance(c *gin.Context) {
	response, err := client.GetBalances(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
	})
}

func item(c *gin.Context) {
	response, err := client.GetItem(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	institution, err := client.GetInstitutionByID(response.Item.InstitutionID)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item":        response.Item,
		"institution": institution.Institution,
	})
}

func identity(c *gin.Context) {
	response, err := client.GetIdentity(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity": response.Accounts,
	})
}

func transactions(c *gin.Context) {
	// pull transactions for the past 30 days
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	response, err := client.GetTransactions(accessToken, startDate, endDate)

	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":     response.Accounts,
		"transactions": response.Transactions,
	})
}

// This functionality is only relevant for the UK Payment Initiation product.
// Retrieve Payment for a specified Payment ID
func payment(c *gin.Context) {
	response, err := client.GetPayment(paymentID)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment": response.Payment,
	})
}

func investmentTransactions(c *gin.Context) {
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")
	response, err := client.GetInvestmentTransactions(accessToken, startDate, endDate)
	fmt.Println("error", err)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"investment_transactions": response,
	})
}

func holdings(c *gin.Context) {
	response, err := client.GetHoldings(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings": response,
	})
}

func info(context *gin.Context) {
	context.JSON(200, map[string]interface{}{
		"item_id":      itemID,
		"access_token": accessToken,
		"products":     strings.Split(PLAID_PRODUCTS, ","),
	})
}

func createPublicToken(c *gin.Context) {
	// Create a one-time use public_token for the Item.
	// This public_token can be used to initialize Link in update mode for a user
	publicToken, err := client.CreatePublicToken(accessToken)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"public_token": publicToken,
	})
}

func createLinkToken(c *gin.Context) {
	linkToken, err := linkTokenCreate(nil)
	if err != nil {
		renderError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"link_token": linkToken})
}

func test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Test success"})
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
	countryCodes := strings.Split(PLAID_COUNTRY_CODES, ",")
	products := strings.Split(PLAID_PRODUCTS, ",")
	redirectURI := PLAID_REDIRECT_URI
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
	resp, err := client.CreateLinkToken(configs)
	if err != nil {
		return "", &httpError{
			errorCode: http.StatusBadRequest,
			error:     err.Error(),
		}
	}
	return resp.LinkToken, nil
}

func assets(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "unfortunate the go client library does not support assets report creation yet."})
}

func respondWithError(c *gin.Context, code int, message interface{}) {
	c.AbortWithStatusJSON(code, gin.H{"error": message})
}

func TokenAuthMiddleware() gin.HandlerFunc {

	// We want to make sure the token is set, bail if not
	// if requiredToken == "" {
	//   log.Fatal("Please set API_TOKEN environment variable")
	// }

	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			respondWithError(c, 403, "API token required")
			return
		}

		httpClient := &http.Client{}
		oauth2Service, err := oauth2.New(httpClient)
		tokenInfoCall := oauth2Service.Tokeninfo()
		tokenInfoCall.AccessToken(token)
		tokenInfo, err := tokenInfoCall.Do()
		if err != nil {
			fmt.Println(err)
			respondWithError(c, 401, "Invalid API token")
			return
		}
		fmt.Println(tokenInfo.HTTPStatusCode)

		c.Next()
	}
}
func main() {
	if APP_PORT == "" {
		APP_PORT = "8000"
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	r.Static("/static", "../static")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api")
	api.Use(TokenAuthMiddleware())
	api.POST("/info", info)
	api.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	// For OAuth flows, the process looks as follows.
	// 1. Create a link token with the redirectURI (as white listed at https://dashboard.plaid.com/team).
	// 2. Once the flow succeeds, Plaid Link will redirect to redirectURI with
	// additional parameters (as required by OAuth standards and Plaid).
	// 3. Re-initialize with the link token (from step 1) and the full received redirect URI
	// from step 2.
	api.GET("/oauth-response.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "oauth-response.html", gin.H{})
	})

	api.POST("/set_access_token", getAccessToken)
	api.POST("/create_link_token_for_payment", createLinkTokenForPayment)
	api.GET("/auth", auth)
	api.GET("/accounts", accounts)
	api.GET("/balance", balance)
	api.GET("/item", item)
	api.POST("/item", item)
	api.GET("/identity", identity)
	api.GET("/transactions", transactions)
	api.POST("/transactions", transactions)
	api.GET("/payment", payment)
	api.GET("/create_public_token", createPublicToken)
	api.POST("/create_link_token", createLinkToken)
	api.GET("/investment_transactions", investmentTransactions)
	api.GET("/holdings", holdings)
	api.GET("/assets", assets)
	api.GET("/categories", getCategories)
	api.GET("/test", test)

	err := r.Run(":" + APP_PORT)
	if err != nil {
		panic("unable to start server")
	}
}

// CHASE ACCESS TOKEN

// {
// 	"access_token": "access-development-1e936067-accd-4af9-a21c-73fcf2b8f1fd",
// 	"item_id": "LvpqB74ad9tP15g1JPDnubVj5ONx4BFZBNZJo"
// }
