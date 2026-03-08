package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"opendi.org/model-hub/api/internal/models/hub"
)

// TODO: Implement actual authentication logic here
// Currently returns a placeholder error
func GetCurrentUser(c *gin.Context) (*hub.User, error) {
	return nil, errors.New("authentication not yet implemented")
}

// OptionalGetCurrentUser is like GetCurrentUser but returns nil (not an error)
// if the user is not authenticated. Used for endpoints that allow both
// authenticated and unauthenticated access
//
// TODO: Implement to match GetCurrentUser logic
func OptionalGetCurrentUser(c *gin.Context) *hub.User {
	user, err := GetCurrentUser(c)
	if err != nil {
		return nil
	}
	return user
}
