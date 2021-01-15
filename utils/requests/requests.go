package requests

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/lakshay35/finlit-backend/models"
)

// ThrowError ...
// Sets error context
func ThrowError(c *gin.Context, code int, reason string) {
	c.AbortWithStatusJSON(code, gin.H{
		"error":  true,
		"reason": reason,
	})
}

// ParseBody ...
// Parses body to defined type
// Throws error if body does not match
func ParseBody(c *gin.Context, res *interface{}) error {
	jsonData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonData, &res)

	if err != nil {
		ThrowError(c, 400, "request body structure match error")
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
