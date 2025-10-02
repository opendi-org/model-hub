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

//note - we technically don't need these structs for now. However, they could be useful in the future.

// ModelHandler struct for handling model requests
type ModelHandler struct {
}

// CommitHandler struct for handling commit requests
type CommitHandler struct {
}

// AuthHandler struct for handling user login/auth requests
type AuthHandler struct {
	googleConfig *oauth2.Config
}

// method for getting an instance of ModelHandler
func NewModelHandler() (*ModelHandler, error) {
	return &ModelHandler{}, nil
}

// method for getting an instance of CommitHandler
func NewCommitHandler() (*CommitHandler, error) {
	return &CommitHandler{}, nil
}

// method for getting an instance of AuthHandler
func NewAuthHandler(id, secret string) (*AuthHandler, error) {
	return &AuthHandler{
		googleConfig: &oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}, nil
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
	_, user, err := database.GetUserByID(userID)
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
// @Description  Processes Callback and returns user info
// @Tags         auth
// @Param        code   query  string  true  "Authorization code from Google"
// @Param        state  query  string  true  "Random state token from login"
// @Success      200 {object} gin.H "Success with body containing user info"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      500 {object} gin.H "Internal Server Error"
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

	// find or create user based on the user info
	user, err := database.FindOrCreateUserFromGoogle(userInfo["name"].(string), userInfo["email"].(string), userInfo["id"].(string))
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

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": jwtToken,
	})
}

// GetModelPrivacy godoc
// @Summary      Get privacy settings for model
// @Description  get privacy settings for model
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      401
// @Failure      403
// @Failure      500
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
	_, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	resp := gin.H{
		"isPublic": model.IsPublic,
		"shares":   model.Shares,
	}

	c.JSON(http.StatusOK, resp)
}

// PutModelPrivacy godoc
// @Summary      Get privacy settings for model
// @Description  get privacy settings for model
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
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
	_, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.OwnerID != actingUserID {
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
		_, _, err := database.GetUserByID(share.UserID)
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
// @Description  get ownership transfer request for model
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      401
// @Failure      403
// @Failure      404
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
// @Description  create ownership transfer request for model
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      400
// @Failure      403
// @Failure      404
// @Failure      500
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
	status, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// make sure the requesting user owns the model
	if model.OwnerID != actingUserID {
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
// @Description  accept/decline ownership transfer request for model
// @Tags         models
// @Success      200
// @Failure      400
// @Failure      403
// @Failure      404
// @Failure      500
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
// @Description  gets all models
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/models/ [get]
func (h *ModelHandler) GetModels(c *gin.Context) {
	status, models, err := database.GetAllModels()
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(status, models)
}

// UploadModel godoc
// @Summary      Upload a new model
// @Description  Given a body of a model with a creator with an email that corresponds to a user in the database, creates the model.
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.CausalDecisionModel  true  "Causal Decision Model Payload"
// @Success      201 {object} apiTypes.CausalDecisionModel "Created model"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      409 {object} gin.H "Conflict: Model with same UUID already exists"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	// Bind the JSON payload to the uploaded model struct
	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the encapsulated CreateModel method from the database package
	if status, err := database.CreateModel(&uploadedModel); err != nil {
		// Return error based on the CreateModel function response
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// Return a successful response if model creation is successful
	c.JSON(http.StatusCreated, uploadedModel)
}

// GetModelByUUID godoc
// @Summary      Get model by its uuid
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Router       /v0/models/{uuid} [get]
func (h *ModelHandler) GetModelByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	status, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// Return the model if found
	c.JSON(status, model)
}

// putModel godoc
// @Summary      Update model
// @Description  Updates a causal decision model along with its metadata in a single transaction.
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.CausalDecisionModel  true  "Causal Decision Model Payload"
// @Success      201 {object} apiTypes.CausalDecisionModel "Updated model"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [put]
func (h *ModelHandler) PutModel(c *gin.Context) {

	var uploadedModel apiTypes.CausalDecisionModel

	// Bind the JSON payload to the uploaded model struct
	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	//if we can't find the model with the given UUID, return error.
	status, oldmodel, err := database.GetModelByUUID(uploadedModel.Meta.UUID)

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	changedModel, status, err := database.UpdateModelAndCreateCommit(&uploadedModel, oldmodel)
	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	// Return a successful response if model put is
	c.JSON(http.StatusCreated, changedModel)
}

// GetCommits godoc
// @Summary      Get all commits
// @Description  gets all commits
// @Tags         commits
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/commits/ [get]
func (h *CommitHandler) GetCommits(c *gin.Context) {
	//TODO remove this API. No real need for it.
	status, models, err := database.GetAllCommits()
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(status, models)
}

func (h *CommitHandler) GetLatestCommitByModelUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	status, commit, err := database.GetLatestCommitForModelUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the commit if found
	c.JSON(status, commit)
}

// doesn't do anything to the database, but just returns the version of the model associated with the commit version

// GetVersionOfModel godoc
// @Summary      Get version of model
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Param        version path string true "Model Version"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/version/{uuid}/{version} [get]
func (h *ModelHandler) GetVersionOfModel(c *gin.Context) {
	strVersion := c.Param("version")
	uuid := c.Param("uuid")
	version, err := strconv.Atoi(strVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	//get latest version of model.
	_, latestVersionOfModel, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	//get latest commit for model UUID.
	status, commit, err := database.GetLatestCommitForModelUUID(uuid)
	if version == 0 && status == http.StatusNotFound {
		c.JSON(http.StatusOK, latestVersionOfModel)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	//if the version requested is the latest version of the model, just return.
	if version == commit.Version {
		c.JSON(http.StatusOK, latestVersionOfModel)
		return
	}
	//if the version requested is greater than the latest version of the model, return error.
	if version > commit.Version {
		c.JSON(http.StatusConflict, gin.H{"Error": "Version requested is greater than the latest version"})
		return
	}
	if version < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Version requested is less than 0"})
		return
	}

	currVersion := commit.Version
	currCommit := commit

	currModelBytes, err := json.Marshal(latestVersionOfModel)

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// loop through the commits until we reach the version we want.
	// we need to apply the diff to the model in reverse order, so we start with the latest commit and go backwards.
	for {
		diff := []byte(currCommit.Diff)
		modified, err := jsonDiffHelpers.ApplyInvertedPatch(currModelBytes, diff)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//reset variables for next iteration of applying patches
		currModelBytes = modified
		currVersion--
		parentIdStr := currCommit.ParentCommitID

		//if we reach the version we want, return the model
		if currVersion <= version {
			break
		}

		if parentIdStr == "" {
			c.JSON(http.StatusInternalServerError, "No parent ID") //if we encounter a null parent id, return error.
			return
		}

		parentId, _ := strconv.ParseInt(parentIdStr, 10, 64)

		_, currCommit, _ = database.GetCommitByID(int(parentId))

	}

	finalModel := apiTypes.CausalDecisionModel{}
	if err := json.Unmarshal(currModelBytes, &finalModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, finalModel)

}

/* //ERIC - we only needed this for testing.
// UploadCommit godoc
// @Summary      Upload a new commit
// @Description  Uploads a commit
// @Tags         commits
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.Commit  true  "Commit Payload"
// @Success      201 {object} apiTypes.Commit "Created Commit"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      409 {object} gin.H "Conflict: Commit with same UUID already exists"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/commits/ [post]
func (h *CommitHandler) UploadCommit(c *gin.Context) {
	var uploadedCommit apiTypes.Commit

	// Bind the JSON payload to the uploaded commit struct
	if err := c.ShouldBindJSON(&uploadedCommit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Call the encapsulated CreateCommit method from the database package
	if status, err := database.CreateCommit(&uploadedCommit); err != nil {
		// Return error based on the CreateCommit function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response if commit creation is successful
	c.JSON(http.StatusCreated, uploadedCommit)
}
*/

// GetModelLineage godoc
// @Summary      Get model lineage
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Router       /v0/models/lineage/{uuid} [get]

func (h *ModelHandler) GetModelLineage(c *gin.Context) {
	uuid := c.Param("uuid")
	status, lineage, err := database.GetModelLineage(uuid)
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(status, lineage)
}

// GetModelChildren godoc
// @Summary      Get model children
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Router       /v0/models/children/{uuid} [get]
func (h *ModelHandler) GetModelChildren(c *gin.Context) {
	uuid := c.Param("uuid")
	status, children, err := database.GetModelChildren(uuid)
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(status, children)
}

// ModelSearch godoc
// @Summary      Search for models
// @Description  Search for models by name or user
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        type path string true "Search type (model or user)"
// @Param        name path string true "Search name"
// @Success      200 {object} []apiTypes.CausalDecisionModel "List of models"
// @Failure      404 {object} gin.H "Model not found"
// @Failure      500 {object} gin.H "Internal Server Error"
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
		c.JSON(status, models)
	case "user":
		status, models, err := database.SearchModelsByUser(name)
		if err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(status, models)
	default:
		c.JSON(404, gin.H{"error": "This type of search does not exist"})
		return
	}
}

// GetCommitsByModelUUID godoc
// @Summary      Get all commits for a model
// @Description  gets all commits for a specific model by its UUID
// @Tags         commits
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Commits not found"
// @Router       /v0/commits/model/{uuid} [get]
func (h *CommitHandler) GetCommitsByModelUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the database function to get all commits for the model
	status, commits, err := database.GetCommitsByModelUUID(uuid)
	if err != nil {
		// If error, return an appropriate response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the commits if found
	c.JSON(status, commits)
}
