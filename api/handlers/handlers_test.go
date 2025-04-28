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
		models.GET("/modelVersion/:uuid/:version", modelHandler.GetVersionOfModel)
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
	//we can log the user in TODO - is this true? someone check on this later.
	req1, _ := http.NewRequest("POST", "/login?email=creator@example.com&password=pass1", nil)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	//test creating a new model.
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
	database.CreateExampleModels()
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

	// try a type of search that doesnt exist
	req3, _ := http.NewRequest("GET", "/v0/models/search/fake/summary", nil)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var responseBody3 []map[string]interface{}
	err3 := json.Unmarshal(w3.Body.Bytes(), &responseBody3)
	assert.Error(t, err3)
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

	//update the example model with the updated example model.
	reqBody := bytes.NewBuffer(example)
	req, _ := http.NewRequest("PUT", "/v0/models", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// try to update with a Nil - should return bad request
	req2, _ := http.NewRequest("PUT", "/v0/models", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// try to update with a model currently not in the database.
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

	example2, err := os.ReadFile("../test_files/updatedExampleModel2.json")
	if err != nil {
		t.Errorf("Error reading test data: %s", err)

	}

	//Need to have the user be created in order for this to work, so
	//we can log the user in
	req1, _ := http.NewRequest("POST", "/login?email=creator@example.com&password=pass1", nil)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

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
	database.CreateExampleModels()

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

	_, model, _ := database.GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
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
