package main

import (
	"fmt"
	"net/http"
	"strings"

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
			ThrowError(c, 403, "API token required")
			return
		}

		httpClient := &http.Client{}
		oauth2Service, err := oauth2.New(httpClient)
		tokenInfoCall := oauth2Service.Tokeninfo()
		tokenInfoCall.AccessToken(token)
		tokenInfo, err := tokenInfoCall.Do()

		if err != nil {
			fmt.Println(err)
			ThrowError(c, 401, "Invalid API token")
			return
		}

		// Bypasses user enrichment when a user is being registered
		if !strings.HasSuffix(c.Request.RequestURI, "/user/register") && !strings.HasSuffix(c.Request.RequestURI, "/user/profile") {
			user, err := GetUser(tokenInfo.UserId)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "User not registered. Make sure you register at /user/register before you make any other requests",
					"error":   true,
				})

				return
			}

			c.Set("USER", *user)
		} else {
			c.Set("USERID", tokenInfo.UserId)
		}

		c.Next()
	}
}
