package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/option"

	"github.com/gin-gonic/gin"
	userService "github.com/lakshay35/finlit-backend/services/user"
	"github.com/lakshay35/finlit-backend/utils/requests"
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
			requests.ThrowError(c, 403, "API token required. Pass in 'Authorization' header")
			return
		}

		oauth2Service, err := oauth2.NewService(context.Background(), option.WithHTTPClient(&http.Client{}))

		if err != nil {
			panic(err)
		}

		tokenInfoCall := oauth2Service.Tokeninfo()
		tokenInfoCall.AccessToken(token)
		tokenInfo, err := tokenInfoCall.Do()

		if err != nil {
			fmt.Println(err)
			requests.ThrowError(c, 401, "Invalid API token")
			return
		}

		// Bypasses user enrichment when a user is being registered
		if !strings.HasSuffix(c.Request.RequestURI, "/user/register") && !strings.HasSuffix(c.Request.RequestURI, "/user/profile") {
			user, err := userService.GetUser(tokenInfo.UserId)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"reason": "User not registered. Make sure you register at /user/register before you make any other requests",
					"error":  true,
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
