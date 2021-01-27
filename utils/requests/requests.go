package requests

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
)

// ThrowError ...
// Sets error context
func ThrowError(c *gin.Context, code int, reason string) {

	// TODO: Add log statement here

	c.AbortWithStatusJSON(code, gin.H{
		"error":  true,
		"reason": reason,
	})
}

// ParseBody ...
// Parses body to defined type
// Throws error if body does not match
func ParseBody(c *gin.Context, res interface{}) error {
	err := c.BindJSON(&res)

	if err != nil {
		fmt.Println(err.Error())
		ThrowError(c, 400, "Request body deserialization error")
		return err
	}

	return nil
}

// GetUserFromContext ...
// Returns user object form context
func GetUserFromContext(c *gin.Context) models.User {
	user, exists := c.Get("USER")

	if !exists {
		panic("USER DOES NOT EXIST IN CONTEXT")
	}

	return user.(models.User)
}

// GetUserIDFromContext ...
// Returns user object form context
func GetUserIDFromContext(c *gin.Context) string {
	user, exists := c.Get("USERID")

	if !exists {
		panic("USERID DOES NOT EXIST IN CONTEXT")
	}

	return user.(string)
}

// ParseHeaders ...
// Parses list of headers from request
// func ParseHeaders(c *gin.Context, headers ...string) ([]string, *errors.Error) {

// 	for i, header := range headers {
// 		budgetID := c.GetHeader(header)
// 	}
// }
