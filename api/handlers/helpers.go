//
// COPYRIGHT OpenDI
//

package handlers

import (
	"encoding/json"
	"fmt"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getUserIDFromToken(c *gin.Context, engMode bool) (int, error) {
	if engMode {
		return 0, nil
	}

	var tokenString string

	// get authorization header and extract token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return -1, fmt.Errorf("invalid authorization format")
		}
	} else {
		// Fallback: Try to get from Cookie (Standard for Google OAuth flow)
		cookie, err := c.Cookie("auth_token")
		if err != nil {
			return -1, fmt.Errorf("authorization header or cookie required")
		}
		tokenString = cookie
	}

	// get the secret from environment
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		return -1, fmt.Errorf("environment variable JWT_SECRET is not set or empty")
	}

	// parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return -1, fmt.Errorf("invalid or expired token")
	}

	// extract claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, fmt.Errorf("invalid or expired token")
	}

	// get and return the userID from the token
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return -1, fmt.Errorf("user_id not found in token")
	}
	return int(userID), nil
}

func addAddonsFields(model apiTypes.CausalDecisionModel) gin.H {
	jsonData, _ := json.Marshal(model)
	var result gin.H
	json.Unmarshal(jsonData, &result)

	result["addons"] = gin.H{
		"ownerID":  model.Addons.OwnerID,
		"tag":      model.Addons.Tag,
		"isPublic": model.Addons.IsPublic,
	}

	return result
}

func filterModelsForActingUser(actingUserID int, models []apiTypes.CausalDecisionModel) []apiTypes.CausalDecisionModel {
	// get user if authenticated
	var actingUser *apiTypes.User
	if actingUserID != -1 {
		actingUser, _ = database.GetUserByID(actingUserID)
	}

	// filter models based on permissions
	filteredModels := make([]apiTypes.CausalDecisionModel, 0)
	for _, model := range models {
		if model.Addons.IsPublic {
			filteredModels = append(filteredModels, model)
			continue
		}

		if actingUserID == -1 {
			continue
		}

		if model.Addons.OwnerID == actingUserID {
			filteredModels = append(filteredModels, model)
			continue
		}

		hasAccess := false
		for _, share := range model.Addons.Shares {
			if share.Email == actingUser.Email {
				hasAccess = true
				break
			}
		}
		if hasAccess {
			filteredModels = append(filteredModels, model)
		}
	}

	return filteredModels
}

func checkIfUserHasAccessToModel(actingUserID int, model apiTypes.CausalDecisionModel) bool {
	hasAccess := model.Addons.IsPublic || model.Addons.OwnerID == actingUserID
	if actingUserID != -1 && !hasAccess {
		actingUser, _ := database.GetUserByID(actingUserID)
		for _, share := range model.Addons.Shares {
			if share.Email == actingUser.Email {
				hasAccess = true
				break
			}
		}
	}
	return hasAccess
}
