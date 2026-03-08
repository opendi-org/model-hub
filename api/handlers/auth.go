//
// COPYRIGHT OpenDI
//

package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"opendi/model-hub/api/database"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthHandler struct for handling user login/auth requests
type AuthHandler struct {
	googleConfig *oauth2.Config
}

// NewAuthHandler method for getting an instance of AuthHandler
func NewAuthHandler(id, secret string) *AuthHandler {
	return &AuthHandler{
		googleConfig: &oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  "http://localhost:3000/auth/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GoogleConfig returns the OAuth2 config for testing purposes
func (h *AuthHandler) GoogleConfig() *oauth2.Config {
	return h.googleConfig
}

// @Router /auth/testlogin [get]
func (h *AuthHandler) TestLogin(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get the user from database
	user, err := database.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// create a token for this user
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "environment variable JWT_SECRET is not set or empty"})
		return
	}
	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// set the cookie and the token
	c.SetCookie("auth_token", jwtToken, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": jwtToken,
	})
}

// GoogleLogin godoc
// @Summary      Start Google OAuth flow
// @Description  Redirects user to Google OAuth consent screen
// @Tags         auth
// @Success      302
// @Failure      500
// @Router       /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	state := hex.EncodeToString(stateBytes)
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // secure is false here so we can use on localhost

	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary      Handle Google OAuth callback
// @Description  Processes callback and returns user info
// @Tags         auth
// @Param        code   query  string  true  "Authorization code from Google"
// @Param        state  query  string  true  "State token for validation"
// @Success      200    {object}  gin.H  "User info"
// @Failure      400    {object}  gin.H  "Invalid state token or missing parameters"
// @Failure      500    {object}  gin.H  "Internal server error"
// @Router       /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// make sure state token matches what the login set
	storedState, err := c.Cookie("oauth_state")
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state token"})
		return
	}

	// exchange the authorization code for an access token
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// make request to google to get the user info
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// decode the user info
	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	picture := ""
	if pic, ok := userInfo["picture"].(string); ok {
		picture = pic
	}

	// find or create user based on the user info
	user, err := database.FindOrCreateUserFromGoogle(userInfo["name"].(string), userInfo["email"].(string), userInfo["id"].(string), picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create a token for this user
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "environment variable JWT_SECRET is not set or empty"})
		return
	}
	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set JWT as httpOnly cookie
	c.SetCookie("auth_token", jwtToken, 3600*24, "/", "localhost", false, true)
	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged in successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"picture":  user.Picture,
		},
		"token": jwtToken,
	})
}

// GetCurrentUser godoc
// @Summary      Get current authenticated user
// @Description  Verifies JWT token and returns current user info
// @Tags         auth
// @Produce      json
// @Success      200 {object} apiTypes.User
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      404 {object} gin.H "User not found"
// @Router       /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	cookie, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not configured"})
		return
	}

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		return
	}

	user, err := database.GetUserByID(int(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return user data along with the token
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"picture":  user.Picture,
		"token":    cookie,
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Clears authentication cookie
// @Tags         auth
// @Success      200 {object} gin.H
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
