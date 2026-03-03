//
// COPYRIGHT OpenDI
//

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

// TestMain is the entry point for the test suite. It sets up the test environment and runs the tests.
func TestMain(m *testing.M) {
	// setup test env
	setup()
	os.Exit(m.Run())
}

// setup initializes the test environment by loading environment variables and setting up the database.
func setup() {

	//import environment variables
	err := godotenv.Load("../config/.env.test")
	if err != nil {
		fmt.Println("Error importing environment variables: ", err)
		os.Exit(1)
	}
	ret := 0
	//we also test the initialize DB instance here
	ret, err = database.InitializeDBInstance()
	if ret != 0 {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}

	database.ResetTables()

	// Initialize router
	router = SetUpRouter()

}

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	//initialize handlers
	modelHandler := NewModelHandler(false)

	authHandler := NewAuthHandler("test-client-id", "test-client-secret")

	//router group for all endpoints related to models
	models := r.Group("/v0/models")
	{
		models.GET("", modelHandler.GetModels)
		models.POST("", modelHandler.UploadModel)
		models.GET("/:uuid", modelHandler.GetModelByUUID)
		models.GET("/tag/:tag", modelHandler.GetModelByTag)
		models.GET("/commits/:tag", modelHandler.GetCommitsByModelTag)
		models.GET("/commits/latest/:tag", modelHandler.GetLatestCommitByModelTag)
		models.GET("/lineage/:tag", modelHandler.GetModelLineage)
		models.GET("/children/:tag", modelHandler.GetModelChildren)
		models.GET("/version/:uuid/:version", modelHandler.GetVersionOfModel)
		models.GET("/search/:type/:name", modelHandler.ModelSearch)
		models.GET("/privacy/:tag", modelHandler.GetModelPrivacy)
		models.PUT("/privacy/:tag", modelHandler.PutModelPrivacy)
		models.GET("/transfer/:tag", modelHandler.GetTransfer)
		models.POST("/transfer/:tag", modelHandler.PostTransfer)
		models.DELETE("/transfer/:tag", modelHandler.DeleteTransfer)
	}

	auth := r.Group("/auth")
	{
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/testlogin", authHandler.TestLogin)
		auth.GET("/me", authHandler.GetCurrentUser)
		auth.POST("/logout", authHandler.Logout)
	}

	return r
}

func getTestToken(id int) string {
	loginReq, _ := http.NewRequest("GET", fmt.Sprintf("/auth/testlogin?id=%d", id), nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	var loginResp map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	return loginResp["token"].(string)
}

func TestGetModels(t *testing.T) {
	database.ResetTables()
	req, _ := http.NewRequest("GET", "/v0/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetModelByUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	req, _ := http.NewRequest("GET", "/v0/models/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	req, _ = http.NewRequest("GET", "/v0/models/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestGetModelByTag(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	req, _ := http.NewRequest("GET", "/v0/models/tag/test-model:1.0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Model", response.Meta.Name)
}

func TestUploadModelNoUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	meta := apiTypes.Meta{
		Name:    "New Model",
		Summary: "This is a test model",
		Version: "1.0",
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resultModel apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &resultModel)
	assert.NotEmpty(t, resultModel.Meta.UUID)

	// should not be any commits
	getReq, _ := http.NewRequest("GET", "/v0/models/commits/new-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)

	// check the latest commit, which should not exist
	getReq, _ = http.NewRequest("GET", "/v0/models/commits/latest/new-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW = httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)
}

func TestUploadModelBadUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	// bad UUID that we did not create, should be overridden
	meta := apiTypes.Meta{
		UUID:    "2a3b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6m",
		Name:    "New Model",
		Summary: "This is a test model",
		Version: "1.0",
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resultModel apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &resultModel)
	assert.NotEmpty(t, resultModel.Meta.UUID)
	assert.NotEqual(t, meta.UUID, resultModel.Meta.UUID)

	// should not be any commits
	getReq, _ := http.NewRequest("GET", "/v0/models/commits/new-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)

	// check the latest commit, which should not exist
	getReq, _ = http.NewRequest("GET", "/v0/models/commits/latest/new-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW = httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)
}

// also tests GetLatestCommitByModelUUID
func TestUploadModelExistingModel(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	meta := apiTypes.Meta{
		UUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		Name:    "Test Model",
		Summary: "This is an updated test model", // updated the description
		Version: "2.0",                           // updated the version
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resultModel apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &resultModel)
	assert.Equal(t, meta.UUID, resultModel.Meta.UUID)
	assert.Equal(t, "This is an updated test model", resultModel.Meta.Summary)

	// we should have a single commit
	getReq, _ := http.NewRequest("GET", "/v0/models/commits/test-model:2.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var response1 []apiTypes.Commit
	json.Unmarshal(getW.Body.Bytes(), &response1)

	assert.Equal(t, 1, len(response1))
	assert.Equal(t, 1, response1[0].UserID)

	// check the latest commit, which should exist
	getReq, _ = http.NewRequest("GET", "/v0/models/commits/latest/test-model:2.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW = httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var response2 apiTypes.Commit
	json.Unmarshal(getW.Body.Bytes(), &response2)

	assert.Equal(t, 1, response2.UserID)
	assert.NotEmpty(t, response2.Diff)
}

func TestUploadModelExistingModelUnchangedVersion(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	meta := apiTypes.Meta{
		UUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		Name:    "Test Model",
		Summary: "This is an updated test model", // updated the description
		Version: "1.0",                           // did not update the version
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadModelInvalidPermissions(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	database.UpdateModelPrivacyByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", false, []apiTypes.Share{
		{
			Email: "childcreator@gmail.com",
			Level: "read",
		},
	})

	// userID 2 will only have read access
	token := getTestToken(2)

	meta := apiTypes.Meta{
		UUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		Name:    "Test Model",
		Summary: "This is an updated test model",
		Version: "2.0",
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUploadModelUnauthorized(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	meta := apiTypes.Meta{
		UUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		Name:    "Test Model",
		Summary: "This is an updated test model",
		Version: "2.0",
		Draft:   false,
		Creator: apiTypes.User{
			Username: "creator",
			Email:    "creator@gmail.com",
		},
	}

	model := apiTypes.CausalDecisionModel{
		Schema: "Test Schema",
		Meta:   meta,
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/v0/models", bytes.NewBuffer(body))
	// missing authorization
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetModelLineage(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	//tests if the handler returns a 200 OK status code when the model exists for the model lineage
	req, _ := http.NewRequest("GET", "/v0/models/lineage/test-child-model:1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// tests whether we can get the children of a model. This is an OK test given that the route function is just a wrapper for the database function.
func TestGetModelChildren(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	req, _ := http.NewRequest("GET", "/v0/models/children/test-model:1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestModelSearch(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	//First let's search by model name and summary
	req1, _ := http.NewRequest("GET", "/v0/models/search/model/summary", nil)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var responseBody []map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	assert.Equal(t, len(responseBody), 1)
	assert.Equal(t, "Test Child Model", responseBody[0]["meta"].(map[string]interface{})["name"])

	//next let's search by creator name
	req2, _ := http.NewRequest("GET", "/v0/models/search/user/creator", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var responseBody2 []map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &responseBody2)
	assert.NoError(t, err2)

	assert.Equal(t, len(responseBody), 1)
	assert.Contains(t, responseBody2[0]["meta"].(map[string]interface{})["name"], "Test")
	assert.Contains(t, responseBody2[1]["meta"].(map[string]interface{})["creator"].(map[string]interface{})["email"], "creator@gmail.com")

	// try a type of search that doesnt exist
	req3, _ := http.NewRequest("GET", "/v0/models/search/fake/summary", nil)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var responseBody3 []map[string]interface{}
	err3 := json.Unmarshal(w3.Body.Bytes(), &responseBody3)
	assert.Error(t, err3)
}

// tests getting different versions of models
func TestGetVersionOfModel(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// get initial version of model that has not been updated
	req, _ := http.NewRequest("GET", "/v0/models/version/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedModel apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModel)

	currentModel, _ := database.GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Equal(t, currentModel.Meta.Version, returnedModel.Meta.Version)
	assert.Equal(t, currentModel.Meta.Summary, returnedModel.Meta.Summary)

	// get version with non-existent UUID
	req, _ = http.NewRequest("GET", "/v0/models/version/non-existent-uuid/1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// update model and create version 2.0
	returnedModel.Meta.Summary = "Updated summary"
	returnedModel.Meta.Version = "2.0"
	database.UpdateModelAndCreateCommit(&returnedModel, currentModel, 1)

	// get latest version of updated model
	req, _ = http.NewRequest("GET", "/v0/models/version/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/2.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedModelV2 apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModelV2)
	assert.Equal(t, "2.0", returnedModelV2.Meta.Version)
	assert.Equal(t, "Updated summary", returnedModelV2.Meta.Summary)

	// get non-existent version
	req, _ = http.NewRequest("GET", "/v0/models/version/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/3.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// get old version after model has been updated
	req, _ = http.NewRequest("GET", "/v0/models/version/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedModelV1 apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModelV1)
	assert.Equal(t, "1.0", returnedModelV1.Meta.Version)
	assert.Equal(t, currentModel.Meta.Summary, returnedModelV1.Meta.Summary)
	assert.NotEqual(t, "Updated summary", returnedModelV1.Meta.Summary)
}

func TestGoogleLogin(t *testing.T) {
	req, _ := http.NewRequest("GET", "/auth/google/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	// location header should be set correctly
	location := w.Header().Get("Location")
	assert.NotEmpty(t, location)
	assert.Contains(t, location, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, location, "client_id=test-client-id")

	// state cookie should be set correctly
	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauth_state" {
			stateCookie = cookie
			break
		}
	}

	// state cookie should not be empty either
	assert.NotNil(t, stateCookie)
	assert.NotEmpty(t, stateCookie.Value)
}

func TestGoogleCallbackInvalidState(t *testing.T) {
	req, _ := http.NewRequest("GET", "/auth/google/callback?code=test-code&state=wrong-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: "badstate",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoogleCallbackMissingState(t *testing.T) {
	// request is missing state cookie
	req, _ := http.NewRequest("GET", "/auth/google/callback?code=test-code&state=some-state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPutAndGetModelPrivacy(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	type privacy struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}

	putBody := privacy{
		IsPublic: true,
		Shares: []apiTypes.Share{
			{
				Email: "childcreator@gmail.com",
				Level: "write",
			},
		},
	}
	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/test-model:1.0", bytes.NewBuffer(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	// should be updated now
	assert.Equal(t, http.StatusOK, putW.Code)

	// get the privacy settings to check
	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/test-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var response privacy
	json.Unmarshal(getW.Body.Bytes(), &response)

	assert.Equal(t, true, response.IsPublic)
	assert.Equal(t, 1, len(response.Shares))
	assert.Equal(t, "childcreator@gmail.com", response.Shares[0].Email)
	assert.Equal(t, "write", response.Shares[0].Level)
}

func TestPutModelPrivacyInvalidPermissions(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// userID 2 does not have owner access to this model
	token := getTestToken(2)

	type privacy struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}

	putBody := privacy{
		IsPublic: true,
		Shares:   []apiTypes.Share{},
	}

	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/test-model:1.0", bytes.NewReader(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	assert.Equal(t, http.StatusForbidden, putW.Code)
}

func TestPutModelPrivacyInvalidTag(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	type privacy struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}

	putBody := privacy{
		IsPublic: true,
		Shares:   []apiTypes.Share{},
	}

	// bad tag for request
	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/badtag", bytes.NewReader(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	assert.Equal(t, http.StatusNotFound, putW.Code)
}

func TestPutModelPrivacyInvalidShares(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	type privacy struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}

	// shouldn't be able to set public to true AND have a read level access
	putBody := privacy{
		IsPublic: true,
		Shares: []apiTypes.Share{
			{
				Email: "childcreator@gmail.com",
				Level: "read",
			},
		},
	}

	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/test-model:1.0", bytes.NewReader(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	assert.Equal(t, http.StatusBadRequest, putW.Code)
}

func TestGetModelPrivacyInvalidPermissions(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// userID 2 does not have owner access to this model
	token := getTestToken(2)

	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/test-model:1.0", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusForbidden, getW.Code)
}

func TestGetModelPrivacyInvalidTag(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	// bad tag for request
	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/badtag", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)
}

func TestPostTransfer(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	req, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apiTypes.Transfer
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", response.CDMUUID)
	assert.Equal(t, 1, response.FromUserID)
	assert.Equal(t, 2, response.ToUserID)
}

func TestPostTransferInvalidPermissions(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// user should not be able to create a ownership transfer request for a model they do not own
	token := getTestToken(2)

	req, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=creator@gmail.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPostTransferMissingOwnerParam(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	// missing owner query parameter
	req, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostTransferAlreadyExists(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	// create first transfer
	req1, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// try to create another transfer on the same model
	req2, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestGetTransfer(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token1 := getTestToken(1)
	postReq1, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	postReq1.Header.Set("Authorization", "Bearer "+token1)
	postW1 := httptest.NewRecorder()
	router.ServeHTTP(postW1, postReq1)

	// should be able to see the transfer as the sending user
	getReq1, _ := http.NewRequest("GET", "/v0/models/transfer/test-model:1.0", nil)
	getReq1.Header.Set("Authorization", "Bearer "+token1)
	getW1 := httptest.NewRecorder()
	router.ServeHTTP(getW1, getReq1)

	assert.Equal(t, http.StatusOK, getW1.Code)

	// make sure response is as expected
	var response apiTypes.Transfer
	json.Unmarshal(getW1.Body.Bytes(), &response)
	assert.Equal(t, "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", response.CDMUUID)
	assert.Equal(t, 1, response.FromUserID)
	assert.Equal(t, 2, response.ToUserID)

	// should be able to see the transfer as the receiving user
	token2 := getTestToken(2)
	getReq2, _ := http.NewRequest("GET", "/v0/models/transfer/test-model:1.0", nil)
	getReq2.Header.Set("Authorization", "Bearer "+token2)
	getW2 := httptest.NewRecorder()
	router.ServeHTTP(getW2, getReq2)

	assert.Equal(t, http.StatusOK, getW2.Code)

	// should not be able to see the transfer as the alternate user
	token3 := getTestToken(3)
	getReq3, _ := http.NewRequest("GET", "/v0/models/transfer/test-model:1.0", nil)
	getReq3.Header.Set("Authorization", "Bearer "+token3)
	getW3 := httptest.NewRecorder()
	router.ServeHTTP(getW3, getReq3)

	assert.Equal(t, http.StatusForbidden, getW3.Code)
}

func TestDeleteTransferAccept(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// user 1 makes ownership transfer request to user 2
	token1 := getTestToken(1)
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 2 accepts transfer
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/test-model:1.0?accept=true", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token2)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusOK, deleteW.Code)

	// transfer should be deleted and owner updated
	transfer, err := database.GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Nil(t, transfer)
	assert.Error(t, err)
	model, _ := database.GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Equal(t, 2, model.Addons.OwnerID)
}

func TestDeleteTransferDecline(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// user 1 makes ownership transfer request to user 2
	token1 := getTestToken(1)
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 2 declines transfer
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/test-model:1.0?accept=false", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token2)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusOK, deleteW.Code)

	// transfer should be deleted and owner should not be updated
	transfer, err := database.GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Nil(t, transfer)
	assert.Error(t, err)
	model, _ := database.GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Equal(t, 1, model.Addons.OwnerID)
}

func TestDeleteTransferInvalidPermissions(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// user 1 makes ownership transfer request to user 2
	token1 := getTestToken(1)
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 1 tries to accept their own transfer request
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/test-model:1.0?accept=true", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token1)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusForbidden, deleteW.Code)
}

func TestDeleteTransferNotFound(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/test-model:1.0?accept=true", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusNotFound, deleteW.Code)
}

func TestDeleteTransferMissingAcceptParam(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	router := SetUpRouter()

	// user 1 makes ownership transfer request to user 2
	token1 := getTestToken(1)
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/test-model:1.0?owner=childcreator@gmail.com", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	// missing accept param
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/test-model:1.0", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token2)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusBadRequest, deleteW.Code)
}

func TestGetCurrentUser(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	token := getTestToken(1)
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response apiTypes.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify user data is returned (ID is not included in JSON response)
	assert.NotEmpty(t, response.Username)
	assert.NotEmpty(t, response.Email)
	assert.Equal(t, "creator", response.Username)
	assert.Equal(t, "creator@gmail.com", response.Email)
}

func TestGetCurrentUserMissingToken(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	req, _ := http.NewRequest("GET", "/auth/me", nil)

	// No cookie added
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Not authenticated")
}

func TestGetCurrentUserInvalidToken(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "invalid.token.string",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid token")
}

func TestGetCurrentUserExpiredToken(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	secret, _ := os.LookupEnv("JWT_SECRET")

	// Create an expired token
	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"email":   "test@test.com",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: expiredToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid token")
}

func TestGetCurrentUserNonExistentUser(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	secret, _ := os.LookupEnv("JWT_SECRET")

	// Create token with non-existent user ID
	nonExistentToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 9999,
		"email":   "nonexistent@test.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: nonExistentToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "User not found")
}

func TestGetCurrentUserMissingUserIDClaim(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	secret, _ := os.LookupEnv("JWT_SECRET")

	// Create token without user_id claim
	invalidClaimsToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@test.com",
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: invalidClaimsToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid user_id in token")
}

func TestLogout(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	token := getTestToken(1)
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Logged out successfully", response["message"])

	// Check that cookie is cleared
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "", authCookie.Value)
	assert.True(t, authCookie.MaxAge < 0)
}

func TestLogoutWithoutCookie(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)

	// No cookie provided
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Logged out successfully", response["message"])

	// Cookie should still be set to clear
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "", authCookie.Value)
}

func TestLogoutCookieProperties(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "auth_token", authCookie.Name)
	assert.Equal(t, "/", authCookie.Path)
	assert.True(t, authCookie.HttpOnly)
	assert.Equal(t, -1, authCookie.MaxAge)
}

func TestGoogleCallbackInvalidAuthCode(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	// Creates a mock Google OAuth server that rejects the auth code
	mockGoogleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/token") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid authorization code",
			})
		}
	}))
	defer mockGoogleServer.Close()

	// Creates a custom AuthHandler with OAuth config pointing to mock server
	testAuthHandler := NewAuthHandler("test-client-id", "test-client-secret")
	testAuthHandler.googleConfig.Endpoint.TokenURL = mockGoogleServer.URL + "/token"

	// Set up a test router
	testRouter := gin.Default()
	testRouter.GET("/auth/google/callback", testAuthHandler.GoogleCallback)

	// Creates request with valid state but invalid code
	state := "test-state-token"
	req, _ := http.NewRequest("GET", "/auth/google/callback?code=invalid-code&state="+state, nil)
	req.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: state,
	})
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// Should return error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["error"])
}
