package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
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
func GetUserFromContext(c *gin.Context) User {
	user, exists := c.Get("USER")

	if !exists {
		panic("USER DOES NOT EXIST IN CONTEXT")
	}

	return user.(User)
}
