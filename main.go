package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/plaid/plaid-go/plaid"
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

	if string(ENCRYPTION_KEY) == "" {
		ENCRYPTION_KEY = []byte("a very very very very secret key") // 32 bytes
	}

	if DATABASE_URL == "" {
		DATABASE_URL = "postgres://fixjakzanvukwk:6f974182090cbbd477d2cc37869c14f0bb685430b358636a27457195cd0c172b@ec2-184-73-249-9.compute-1.amazonaws.com:5432/d21u9s3g6p4143"
	}

}

type account struct {
	UserID      string `json:"accountId"`
	AccessToken string `json:"accessToken"`
}

// Fill with your Plaid API keys - https://dashboard.plaid.com/account/keys
var (
	PLAID_CLIENT_ID     = "5f25d4dbd6e09e001026c0f6"
	PLAID_SECRET        = "690750c7b009518b14ffc6386b7fb4"
	PLAID_ENV           = "development"
	PLAID_PRODUCTS      = "auth,transactions"
	PLAID_COUNTRY_CODES = os.Getenv("PLAID_COUNTRY_CODES")
	ENCRYPTION_KEY      = []byte(os.Getenv("ENCRYPTION_KEY"))
	DATABASE_URL        = os.Getenv("DATABASE_URL")
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

var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// func Encrypt(text string) string {
// 	block, err := aes.NewCipher([]byte(ENCRYPTION_KEY))
// 	if err != nil {
// 		panic(err)
// 	}
// 	plaintext := []byte(text)
// 	cfb := cipher.NewCFBEncrypter(block, iv)
// 	ciphertext := make([]byte, len(plaintext))
// 	cfb.XORKeyStream(ciphertext, plaintext)
// 	return encodeBase64(text)
// }

// func Decrypt(text string) string {
// 	block, err := aes.NewCipher([]byte(ENCRYPTION_KEY))
// 	if err != nil {
// 		panic(err)
// 	}
// 	ciphertext := decodeBase64(text)
// 	cfb := cipher.NewCFBEncrypter(block, iv)
// 	plaintext := make([]byte, len(ciphertext))
// 	cfb.XORKeyStream(plaintext, ciphertext)
// 	return string(ciphertext)
// }

func main() {
	if APP_PORT == "" {
		APP_PORT = "8000"
	}

	InitializeDatabase()

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

	api := r.Group("/api")
	budget := api.Group("/budget")
	user := api.Group("/user")
	role := api.Group("/role")
	account := api.Group("/account")

	api.Use(TokenAuthMiddleware())
	budget.Use(TokenAuthMiddleware())
	user.Use(TokenAuthMiddleware())
	role.Use(TokenAuthMiddleware())
	account.Use(TokenAuthMiddleware())

	// ACCOUNTS
	account.GET("/get", GetAllAccounts)
	account.POST("/get-account-details", GetAccountInformation)
	account.POST("/create_link_token", CreateLinkToken)
	account.POST("/transactions", GetTransactions)
	account.POST("/live-balances", GetCurrentBalances)

	// BUDGETS
	budget.GET("/get", GetBudgets)
	budget.POST("/create", CreateBudget)

	// USERS
	user.POST("/register", RegisterUser)
	user.GET("/profile", GetUserProfile)

	// ROLES
	role.POST("/add-user-role-to-budget", AddUserRoleToBudget)

	err := r.Run(":" + APP_PORT)
	if err != nil {
		panic("unable to start server")
	}
}
