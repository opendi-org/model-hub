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

	"github.com/gin-gonic/gin"
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
	//initialize handler
	modelHandler, err := NewModelHandler()

	authHandler, _ := NewAuthHandler("test-client-id", "test-client-secret")

	commitHandler, _ := NewCommitHandler()

	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	//router group for all endpoints related to commits
	commits := r.Group("/v0/commits")
	{
		commits.GET("", commitHandler.GetCommits) // Get all commits
		commits.GET("/:uuid", commitHandler.GetLatestCommitByModelUUID)
		//commits.POST("", commitHandler.UploadCommit) // Create a commit (for testing)
	}

	//router group for all endpoints related to models
	models := r.Group("/v0/models")
	{
		models.GET("", modelHandler.GetModels)              // Get all models
		models.GET("/:uuid", modelHandler.GetModelByUUID)   // Get a model by UUID
		models.GET("/tag/:tag", modelHandler.GetModelByTag) // Get a model by tag
		models.POST("", modelHandler.UploadModel)           // Update or create a model
		models.GET("/lineage/:uuid", modelHandler.GetModelLineage)
		models.GET("/children/:uuid", modelHandler.GetModelChildren)
		models.GET("/search/:type/:name", modelHandler.ModelSearch)
		models.GET("/modelVersion/:uuid/:version", modelHandler.GetVersionOfModel)
		models.GET("/privacy/:uuid", modelHandler.GetModelPrivacy)
		models.PUT("/privacy/:uuid", modelHandler.PutModelPrivacy)
		models.GET("/transfer/:uuid", modelHandler.GetTransfer)
		models.POST("/transfer/:uuid", modelHandler.PostTransfer)
		models.DELETE("/transfer/:uuid", modelHandler.DeleteTransfer)
	}

	auth := r.Group("/auth")
	{
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/testlogin", authHandler.TestLogin)
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

// TODO: rewrite this test (functionality is now different + need to be authorized)
func TestUploadModel(t *testing.T) {
	database.ResetTables()

	// user must exist to create a model
	database.CreateUser(apiTypes.User{
		Username: "creator",
		Email:    "creator@gmail.com",
		GoogleID: "creator-googleid",
	})

	example, err := os.ReadFile("../test_files/model.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)
	}

	// create a new model
	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("POST", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// tests POST a nil.
	req2, _ := http.NewRequest("POST", "/v0/models", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestGetModelLineage(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()
	//tests if the handler returns a 200 OK status code when the model exists for the model lineage
	req, _ := http.NewRequest("GET", "/v0/models/lineage/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// tests whether we can get the children of a model. This is an OK test given that the route function is just a wrapper for the database function.
func TestGetModelChildren(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	req, _ := http.NewRequest("GET", "/v0/models/children/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
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
	assert.Contains(t, responseBody2[1]["meta"].(map[string]interface{})["creator"].(map[string]interface{})["username"], "childcreator")

	// try a type of search that doesnt exist
	req3, _ := http.NewRequest("GET", "/v0/models/search/fake/summary", nil)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var responseBody3 []map[string]interface{}
	err3 := json.Unmarshal(w3.Body.Bytes(), &responseBody3)
	assert.Error(t, err3)
}

// TODO: rewrite this test (functionality is now different)
func TestGetAllCommits(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	example, err := os.ReadFile("../test_files/updatedExampleModel.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)
	}

	example2, err := os.ReadFile("../test_files/updatedExampleModel2.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)
	}

	//test get all commits  when no models have been updated yet.
	req3, _ := http.NewRequest("GET", "/v0/commits", nil)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.False(t, strings.Contains(w3.Body.String(), "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"))

	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("PUT", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//Test get all commits after a model has been updated.
	req2, _ := http.NewRequest("GET", "/v0/commits", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.True(t, strings.Contains(w2.Body.String(), "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"))

	//test creating a new model does not create a commit[nothing put yet] or break commits
	reqBody6 := bytes.NewBuffer(example2)
	req6, _ := http.NewRequest("POST", "/v0/models", reqBody6)
	req6.Header.Set("Content-Type", "application/json")
	w6 := httptest.NewRecorder()
	router.ServeHTTP(w6, req6)

	assert.Equal(t, http.StatusCreated, w6.Code)

	req4, _ := http.NewRequest("GET", "/v0/commits", nil)
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)
	assert.True(t, strings.Contains(w4.Body.String(), "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"))
	assert.False(t, strings.Contains(w4.Body.String(), "eeee5c4d-5e6f-7eb-140d"))

}

func TestGetLatestCommitByUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	example, err := os.ReadFile("../test_files/updatedExampleModel.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)

	}

	//Need to have the user be created in order for this to work, so
	//we can log the user in
	req1, _ := http.NewRequest("POST", "/login?email=creator@example.com&password=pass1", nil)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	//get the latest commit for the example model.
	req3, _ := http.NewRequest("GET", "/v0/commits/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusNotFound, w3.Code)

	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("PUT", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	//get the latest commit after updating the model.
	req2, _ := http.NewRequest("GET", "/v0/commits/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	//get the latest commit for a model that doesnt exist.
	req4, _ := http.NewRequest("GET", "/v0/commits/fake", nil)
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusNotFound, w4.Code)

}

// tests getting different versions of models.
func TestGetVersionOfModel(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	//tests getting version 0 of a model that has not been updated yet.
	req, _ := http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/0", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var returnedModel apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModel)
	byteReturnedModel, _ := json.Marshal(returnedModel)
	strReturnedModel := string(byteReturnedModel)

	model, _ := database.GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	bytemodel, _ := json.Marshal(model)
	strmodel := string(bytemodel)

	assert.Equal(t, strmodel, strReturnedModel)
	//tests non-number version that results in error.
	req, _ = http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/haha", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	//tests getting model version with a non-existent UUID.
	req, _ = http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4bfff/0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	//push a change to our model.
	returnedModel.Meta.Summary = "Updated summary"
	database.UpdateModelAndCreateCommit(&returnedModel, model)
	//tests getting version 1 of a model that has been updated.
	req, _ = http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/1", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var returnedModel2 apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModel2)
	byteReturnedModel2, _ := json.Marshal(returnedModel2)
	strReturnedModel2 := string(byteReturnedModel2)

	byteReturnedModel, _ = json.Marshal(returnedModel)
	strReturnedModel = string(byteReturnedModel)

	assert.Equal(t, strReturnedModel2, strReturnedModel)
	//tests getting nonexistent version of a model.
	req, _ = http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/2", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	//tests getting version 0 of a model that has been updated.
	req, _ = http.NewRequest("GET", "/v0/models/modelVersion/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d/0", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedModel3 apiTypes.CausalDecisionModel
	json.Unmarshal(w.Body.Bytes(), &returnedModel3)
	byteReturnedModel3, _ := json.Marshal(returnedModel3)
	strReturnedModel3 := string(byteReturnedModel3)

	assert.Equal(t, strReturnedModel3, strmodel)

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
				UserID: 2,
				Level:  "write",
			},
		},
	}
	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", bytes.NewBuffer(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	// should be updated now
	assert.Equal(t, http.StatusOK, putW.Code)

	// get the privacy settings to check
	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var response privacy
	json.Unmarshal(getW.Body.Bytes(), &response)

	assert.Equal(t, true, response.IsPublic)
	assert.Equal(t, 1, len(response.Shares))
	assert.Equal(t, 2, response.Shares[0].UserID)
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
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", bytes.NewReader(jsonBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	router.ServeHTTP(putW, putReq)

	assert.Equal(t, http.StatusForbidden, putW.Code)
}

func TestPutModelPrivacyInvalidUUID(t *testing.T) {
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

	// bad uuid for request
	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/baduuid", bytes.NewReader(jsonBody))
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
				UserID: 2,
				Level:  "read",
			},
		},
	}

	jsonBody, _ := json.Marshal(putBody)
	putReq, _ := http.NewRequest("PUT", "/v0/models/privacy/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", bytes.NewReader(jsonBody))
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

	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusForbidden, getW.Code)
}

func TestGetModelPrivacyInvalidUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	// bad uuid for request
	getReq, _ := http.NewRequest("GET", "/v0/models/privacy/baduuid", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusNotFound, getW.Code)
}

func TestPostTransfer(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(1)

	req, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
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

	req, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=1", nil)
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
	req, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
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
	req1, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// try to create another transfer on the same model
	req2, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestGetTransfer(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token1 := getTestToken(1)
	postReq1, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	postReq1.Header.Set("Authorization", "Bearer "+token1)
	postW1 := httptest.NewRecorder()
	router.ServeHTTP(postW1, postReq1)

	// should be able to see the transfer as the sending user
	getReq1, _ := http.NewRequest("GET", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
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
	getReq2, _ := http.NewRequest("GET", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	getReq2.Header.Set("Authorization", "Bearer "+token2)
	getW2 := httptest.NewRecorder()
	router.ServeHTTP(getW2, getReq2)

	assert.Equal(t, http.StatusOK, getW2.Code)

	// should not be able to see the transfer as the alternate user
	token3 := getTestToken(3)
	getReq3, _ := http.NewRequest("GET", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
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
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 2 accepts transfer
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?accept=true", nil)
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
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 2 declines transfer
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?accept=false", nil)
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
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, http.StatusOK, postW.Code)

	// user 1 tries to accept their own transfer request
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?accept=true", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token1)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusForbidden, deleteW.Code)
}

func TestDeleteTransferNotFound(t *testing.T) {
	database.ResetTables()
	database.CreateExampleData()

	token := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?accept=true", nil)
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
	postReq, _ := http.NewRequest("POST", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d?owner=2", nil)
	postReq.Header.Set("Authorization", "Bearer "+token1)
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	// missing accept param
	token2 := getTestToken(2)
	deleteReq, _ := http.NewRequest("DELETE", "/v0/models/transfer/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token2)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	assert.Equal(t, http.StatusBadRequest, deleteW.Code)
}
