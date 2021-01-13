package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/oauth2/v1"
)

// TokenAuthMiddleware ...Token middleware to validate with google
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

		c.Set("USER", GetUser(tokenInfo.UserId))

		c.Next()
	}
}
