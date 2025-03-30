package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"opendi/model-hub/api/database"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	// setup test env
	setup()
	os.Exit(m.Run())
}

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

	authHandler, _ := NewAuthHandler()

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
		models.GET("", modelHandler.GetModels)            // Get all models
		models.GET("/:uuid", modelHandler.GetModelByUUID) // Get a model by UUID
		models.POST("", modelHandler.UploadModel)         // Upload a model
		models.PUT("", modelHandler.PutModel)             // Update a model
		models.GET("/lineage/:uuid", modelHandler.GetModelLineage)
		models.GET("/children/:uuid", modelHandler.GetModelChildren)
		models.GET("/search/:type/:name", modelHandler.ModelSearch)
	}

	r.POST("/login", authHandler.UserLogin)

	return r
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
	database.CreateExampleModels()
	req, _ := http.NewRequest("GET", "/v0/models/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	req, _ = http.NewRequest("GET", "/v0/models/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestUploadModel(t *testing.T) {
	database.ResetTables()

	example, err := os.ReadFile("../test_files/model.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)

	}

	//Need to have the user be created in order for this to work, so
	//we can log the user in
	req1, _ := http.NewRequest("POST", "/login?email=creator@example.com&password=pass1", nil)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("POST", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// invalid empty(not a model)
	req2, _ := http.NewRequest("POST", "/v0/models", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestGetModelLineage(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

	req, _ := http.NewRequest("GET", "/v0/models/lineage/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetModelChildren(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

	req, _ := http.NewRequest("GET", "/v0/models/children/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestUserLogin(t *testing.T) {
	//Login with a new user
	database.ResetTables()
	req, _ := http.NewRequest("POST", "/login?email=email1&password=pass1", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Parse response body to extract user information
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	// Check that the user email in the response matches the expected one
	assert.Equal(t, "email1", responseBody["email"], "User email should match the login email")
}

func TestModelSearch(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

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
	req2, _ := http.NewRequest("GET", "/v0/models/search/user/test", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var responseBody2 []map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &responseBody2)
	assert.NoError(t, err2)

	assert.Equal(t, len(responseBody), 1)
	assert.Contains(t, responseBody2[0]["meta"].(map[string]interface{})["name"], "Test")
	assert.Contains(t, responseBody2[1]["meta"].(map[string]interface{})["creator"].(map[string]interface{})["username"], "Test")
}

func TestPutModel(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

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

	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("PUT", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// invalid empty(not a model)
	req2, _ := http.NewRequest("PUT", "/v0/models", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// cdm uuid not in database
	model4, err := os.ReadFile("../test_files/model4.json")
	req3Body := bytes.NewBuffer(model4)
	req3, _ := http.NewRequest("PUT", "/v0/models", req3Body)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusNotFound, w3.Code)

}

func TestGetAllCommits(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

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

	req2, _ := http.NewRequest("GET", "/v0/commits", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.True(t, strings.Contains(w2.Body.String(), "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"))
}

func TestGetLatestCommitByUUID(t *testing.T) {
	database.ResetTables()
	database.CreateExampleModels()

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

	req2, _ := http.NewRequest("GET", "/v0/commits/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

}
