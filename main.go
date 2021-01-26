package main

import (
	"time"

	"github.com/lakshay35/finlit-backend/docs"
	services "github.com/lakshay35/finlit-backend/services/environment"

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

func setupSwaggerMetadata() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "FinLit API"
	docs.SwaggerInfo.Description = "This is a REST API for the FinLit Application"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = services.GetEnvVariable("HOST")
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

// @contact.name Lakshay Sharma
// @contact.url sharmalakshay.com
// @contact.email lakshay35@gmail.com

// @query.collection.format multi

// @securityDefinitions.apikey Google AccessToken
// @in header
// @name Authorization

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
			user.GET("/get", routes.GetUserProfile)
		}
		role := api.Group("/role")
		{
			role.POST("/add-user-role-to-budget", routes.AddUserRoleToBudget)
		}
		account := api.Group("/account")
		{
			account.GET("/get", routes.GetAllAccounts)
			account.POST("/get-account-details", routes.GetAccountInformation)
			account.GET("/create-link-token", routes.CreateLinkToken)
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
	// api/user/register should validate body instead of registering a user with empty values
	// api/role/add-user-role-to-budget should be renamed and should validate body instead of returning a 500

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := r.Run(":" + services.GetEnvVariable("PORT"))
	if err != nil {
		panic("unable to start server")
	}
}
