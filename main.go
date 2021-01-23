package main

import (
	"time"

	"github.com/lakshay35/finlit-backend/docs"
	"github.com/lakshay35/finlit-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lakshay35/finlit-backend/docs"
	"github.com/lakshay35/finlit-backend/middlewares"
	"github.com/lakshay35/finlit-backend/routes"
	"github.com/lakshay35/finlit-backend/utils/database"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	// swagger embed files
)

func init() {
	// Initialize main.go
}

type account struct {
	UserID      string `json:"accountId"`
	AccessToken string `json:"accessToken"`
}

func setupSwaggerMetadata() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "FinLit API"
	docs.SwaggerInfo.Description = "This is a REST API for the FinLit Application"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = services.GetEnvVariable("HOST")
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

// var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

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

// @contact.name Lakshay Sharma
// @contact.url sharmalakshay.com
// @contact.email lakshay35@gmail.com

// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information

// @x-extension-openapi {"example": "value on a json format"}
func main() {

	setupSwaggerMetadata()

	database.InitializeDatabase()

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

	// @Failure 403,404 {object} httputil.HTTPError
	api := r.Group("/api")
	{
		api.Use(middlewares.TokenAuthMiddleware())
		budget := api.Group("/budget")
		{
			budget.GET("/get", routes.GetBudgets)
			budget.POST("/create", routes.CreateBudget)
		}
		user := api.Group("/user")
		{
			user.POST("/register", routes.RegisterUser)
			user.GET("/get ", routes.GetUserProfile)
		}
		role := api.Group("/role")
		{
			role.POST("/add-user-role-to-budget", routes.AddUserRoleToBudget)
		}
		account := api.Group("/account")
		{
			account.GET("/get", routes.GetAllAccounts)
			account.POST("/get-account-details", routes.GetAccountInformation)
			account.POST("/create-link-token", routes.CreateLinkToken)
			account.POST("/transactions", routes.GetTransactions)
			account.POST("/live-balances", routes.GetCurrentBalances)
			account.POST("/register-token", routes.RegisterAccessToken)
		}
		expense := api.Group("/expense")
		{
			expense.POST("/add", routes.AddExpense)
			expense.GET("/get", routes.GetAllExpenses)
			expense.GET("/get-expense-charge-cycles", routes.GetExpenseChargeCycles)
			expense.DELETE("/delete/:id", routes.DeleteExpense)
			expense.PUT("/update", routes.UpdateExpense)
		}
	}

	// TODO:
	// api/account/transactions should validate body instead of returning a 500
	// api/account/live-balances should validate body instead of returning a 500
	// api/user/register should validate body instead of registering a user with empty values
	// api/role/add-user-role-to-budget should be renamed and should validate body instead of returning a 500
	// migrate services into services module

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := r.Run(":" + services.GetEnvVariable("PORT"))
	if err != nil {
		panic("unable to start server")
	}
}
