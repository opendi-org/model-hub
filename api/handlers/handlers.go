//
// COPYRIGHT OpenDI
//

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	jsonDiffHelpers "opendi/model-hub/api/jsondiffhelpers"
	"strconv" //for applying patches generated with jsondiff
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	// OAuth 2 Imports
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
	engMode bool
}

// AuthHandler struct for handling user login/auth requests
type AuthHandler struct {
	googleConfig *oauth2.Config
}

// method for getting an instance of ModelHandler
func NewModelHandler(engMode bool) *ModelHandler {
	return &ModelHandler{
		engMode: engMode,
	}
}

// method for getting an instance of AuthHandler
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

func getUserIDFromToken(c *gin.Context) (int, error) {
	// get authorization header and extract token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("authorization header required")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return 0, fmt.Errorf("invalid authorization format")
	}

	// get the secret from environment
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		return 0, fmt.Errorf("environment variable JWT_SECRET is not set or empty")
	}

	// parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid or expired token")
	}

	// extract claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid or expired token")
	}

	// get and return the userID from the token
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token")
	}
	return int(userID), nil
}

func addAddonsFields(model apiTypes.CausalDecisionModel) gin.H {
	jsonData, _ := json.Marshal(model)
	var result gin.H
	json.Unmarshal(jsonData, &result)

	result["addons"] = gin.H{
		"ownerID": model.Addons.OwnerID,
		"tag":     model.Addons.Tag,
	}

	return result
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

	c.JSON(http.StatusOK, user)
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

// GetModelPrivacy godoc
// @Summary      Get privacy settings for model
// @Description  Get privacy settings for model
// @Tags         models
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Success      200   {object}  gin.H  "Privacy settings"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/privacy/{uuid} [get]
func (h *ModelHandler) GetModelPrivacy(c *gin.Context) {
	uuid := c.Param("uuid")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get the model from database
	model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	resp := gin.H{
		"isPublic": model.Addons.IsPublic,
		"shares":   model.Addons.Shares,
	}

	c.JSON(http.StatusOK, resp)
}

// PutModelPrivacy godoc
// @Summary      Update privacy settings for model
// @Description  Update privacy settings for model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Param        privacy  body  object  true  "Privacy settings"
// @Success      200   {object}  gin.H  "Privacy settings updated"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Model not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/privacy/{uuid} [put]
func (h *ModelHandler) PutModelPrivacy(c *gin.Context) {
	uuid := c.Param("uuid")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the model exists
	model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// bind the json to a struct for shares
	var req struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// make sure the shares object is valid
	for _, share := range req.Shares {
		_, err := database.GetUserByID(share.UserID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if (share.Level != "read" && share.Level != "write") || (req.IsPublic && share.Level == "read") {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid share level for userID: %d", share.UserID)})
			return
		}
	}

	// now update the model's privacy settings
	err = database.UpdateModelPrivacyByUUID(uuid, req.IsPublic, req.Shares)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// GetTransfer godoc
// @Summary      Get ownership transfer request for model
// @Description  Get ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Success      200   {object}  gin.H  "Transfer request details"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Transfer request not found"
// @Router       /v0/models/transfer/{uuid} [get]
func (h *ModelHandler) GetTransfer(c *gin.Context) {
	uuid := c.Param("uuid")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// get the transfer
	transfer, err := database.GetTransferByModelUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if transfer.FromUserID != actingUserID && transfer.ToUserID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// PostTransfer godoc
// @Summary      Create ownership transfer request for model
// @Description  Create ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Param        transfer  body  object  true  "Transfer request details"
// @Success      200   {object}  gin.H  "Transfer request created"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Model not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/transfer/{uuid} [post]
func (h *ModelHandler) PostTransfer(c *gin.Context) {
	uuid := c.Param("uuid")
	owner := c.Query("owner")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the owner ID is valid
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner parameter is required"})
		return
	}
	toUserID, err := strconv.Atoi(owner)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner ID"})
		return
	}

	// make sure the model exists
	model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the requesting user owns the model
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// make sure an existing transfer does not exist
	existingTransfer, err := database.GetTransferByModelUUID(uuid)
	if err == nil && existingTransfer != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer request already exists for this model"})
		return
	}

	// create the transfer
	transfer := &apiTypes.Transfer{
		CDMUUID:    uuid,
		ToUserID:   toUserID,
		FromUserID: actingUserID,
		CreatedAt:  time.Now(),
	}
	if err := database.CreateTransfer(transfer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// DeleteTransfer godoc
// @Summary      Accept/decline ownership transfer request for model
// @Description  Accept/decline ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Success      200   {object}  gin.H  "Transfer request processed"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Transfer request not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/transfer/{uuid} [delete]
func (h *ModelHandler) DeleteTransfer(c *gin.Context) {
	uuid := c.Param("uuid")
	accept := c.Query("accept")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the accept parameter is valid
	if accept == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accept parameter is required"})
		return
	}
	acceptBool, err := strconv.ParseBool(accept)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accept parameter must be true or false"})
		return
	}

	// make sure the transfer exists
	transfer, err := database.GetTransferByModelUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the acting user can actually accept this transfer
	if transfer.ToUserID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// delete the transfer
	if err := database.DeleteTransfer(transfer, acceptBool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// GetModels godoc
// @Summary      Get all models
// @Description  Gets all models
// @Tags         models
// @Produce      json
// @Success      200  {array}   gin.H  "List of models"
// @Failure      500  {object}  gin.H  "Internal server error"
// @Router       /v0/models/ [get]
func (h *ModelHandler) GetModels(c *gin.Context) {
	models, err := database.GetAllModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(models))
	for i, model := range models {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// GetModelByUUID godoc
// @Summary      Get model by its UUID
// @Description  Get model by its UUID
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Success      200   {object}  gin.H  "Model"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/{uuid} [get]
func (h *ModelHandler) GetModelByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	model, err := database.GetModelByUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return the model if found
	c.JSON(http.StatusOK, addAddonsFields(*model))
}

// GetModelByTag godoc
// @Summary      Get model by its tag
// @Description  Get model by its tag
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag   path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Model"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/tag/{tag} [get]
func (h *ModelHandler) GetModelByTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return the model if found
	c.JSON(http.StatusOK, addAddonsFields(*model))
}

// UploadModel godoc
// @Summary      Creates or updates a model
// @Description  Creates or updates a model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  object  true  "Model"
// @Success      200    {object}  gin.H  "Model created or updated"
// @Failure      400    {object}  gin.H  "Bad request"
// @Failure      401    {object}  gin.H  "Unauthorized"
// @Failure      403    {object}  gin.H  "Forbidden"
// @Failure      500    {object}  gin.H  "Internal server error"
// @Router       /v0/models [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// if in engine mode, just pretend that acting user is always id 0
	actingUserID := 0
	var err error
	if !h.engMode {
		actingUserID, err = getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
	}

	var resultModel *apiTypes.CausalDecisionModel
	var status int

	if uploadedModel.Meta.UUID != "" {
		// either the model already exists or the user has provided one (which we should ignore and create another one)
		var retrievedModel *apiTypes.CausalDecisionModel
		if retrievedModel, err = database.GetModelByUUID(uploadedModel.Meta.UUID); err != nil {
			// the user provided a UUID that does not match any model in the system, make a new one and ignore the old UUID
			resultModel, status, err = database.CreateModel(&uploadedModel, actingUserID)
		} else {
			// we found a model with a matching UUID, make sure the user has permissions to update
			canWrite := retrievedModel.Addons.OwnerID == actingUserID
			if !canWrite {
				for _, share := range retrievedModel.Addons.Shares {
					if share.UserID == actingUserID && share.Level == "write" {
						canWrite = true
						break
					}
				}
			}
			if !canWrite {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
				return
			}
			// user has permissions to udpate this model, so create a commit
			resultModel, status, err = database.UpdateModelAndCreateCommit(&uploadedModel, retrievedModel, actingUserID)
		}
	} else {
		// a UUID was not provided, we assume this means we are creating a new model
		resultModel, status, err = database.CreateModel(&uploadedModel, actingUserID)
	}
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addAddonsFields(*resultModel))
}

// TODO adjust this so it looks at the version and the first part of the tag separately
// GetVersionOfModel godoc
// @Summary      Get version of model
// @Description  Get version of model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid     path  string  true  "Model UUID"
// @Param        version  path  string  true  "Version number"
// @Success      200      {object}  gin.H  "Model version details"
// @Failure      404      {object}  gin.H  "Version not found"
// @Failure      500      {object}  gin.H  "Internal server error"
// @Router       /v0/models/version/{uuid}/{version} [get]
func (h *ModelHandler) GetVersionOfModel(c *gin.Context) {
	version := c.Param("version")
	uuid := c.Param("uuid")

	// get latest version of model
	latestVersionOfModel, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if latestVersionOfModel.Meta.Version == version {
		c.JSON(http.StatusOK, addAddonsFields(*latestVersionOfModel))
		return
	}

	// get the latest commit; if there isnt one, we are working with a model that has not been updated
	latestCommit, err := database.GetLatestCommitForModelUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	currCommit := latestCommit
	currModelBytes, _ := json.Marshal(latestVersionOfModel)

	// we need to apply the diff to the model in reverse order, so we start with the latest commit and go backwards
	for {
		currModelBytes, err = jsonDiffHelpers.ApplyInvertedPatch(currModelBytes, []byte(currCommit.Diff))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// if we reach the version we want, return the model
		if currCommit.Version == version {
			break
		}
		if currCommit.ParentID == -1 {
			break
		}

		_, currCommit, _ = database.GetCommitByID(currCommit.ParentID)
	}

	finalModel := apiTypes.CausalDecisionModel{}
	if err := json.Unmarshal(currModelBytes, &finalModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if finalModel.Meta.Version != version {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model does not have version: %s", version)})
		return
	}
	c.JSON(http.StatusOK, addAddonsFields(finalModel))

}

// GetModelLineage godoc
// @Summary      Get model lineage
// @Description  Get model lineage
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Model lineage"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/lineage/{tag} [get]
func (h *ModelHandler) GetModelLineage(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	lineage, err := database.GetModelLineage(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(lineage))
	for i, model := range lineage {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// GetModelChildren godoc
// @Summary      Get model children
// @Description  Gets models using its UUID
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {array}   gin.H  "List of child models"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/children/{tag} [get]
func (h *ModelHandler) GetModelChildren(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	children, err := database.GetModelChildren(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(children))
	for i, model := range children {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// ModelSearch godoc
// @Summary      Search for models
// @Description  Search for models by name or user
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        type  path  string  true  "Search type (model or user)"
// @Param        name  path  string  true  "Search name"
// @Success      200   {array}   apiTypes.CausalDecisionModel  "List of models"
// @Failure      404   {object}  gin.H  "Models not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/search/{type}/{name} [get]
func (h *ModelHandler) ModelSearch(c *gin.Context) {
	searchType := c.Param("type")
	name := c.Param("name")
	switch searchType {
	case "model":
		status, models, err := database.SearchModelsByName(name)
		if err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		result := make([]gin.H, len(models))
		for i, model := range models {
			result[i] = addAddonsFields(model)
		}
		c.JSON(status, result)
	case "user":
		status, models, err := database.SearchModelsByUser(name)
		if err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		result := make([]gin.H, len(models))
		for i, model := range models {
			result[i] = addAddonsFields(model)
		}
		c.JSON(status, result)
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "This type of search does not exist"})
		return
	}
}

// GetCommitsByModelTag godoc
// @Summary      Get all commits for a model
// @Description  Get all commits for a model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {array}   gin.H  "List of commits"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/commits/{tag} [get]
func (h *ModelHandler) GetCommitsByModelTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Call the database function to get all commits for the model
	commits, err := database.GetCommitsByModelUUID(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return the commits if found
	c.JSON(http.StatusOK, commits)
}

// GetLatestCommitByModelTag godoc
// @Summary      Get latest commit for a model
// @Description  Get latest commit for a model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Latest commit"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/commits/latest/{tag} [get]
func (h *ModelHandler) GetLatestCommitByModelTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Call the encapsulated GetModelByUUID function from the database package
	commit, err := database.GetLatestCommitForModelUUID(model.Meta.UUID)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}

	// Return the commit if found
	c.JSON(http.StatusOK, commit)
}
