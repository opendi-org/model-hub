//
// COPYRIGHT OpenDI
//

package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	jsonDiffHelpers "opendi/model-hub/api/jsondiffhelpers"
	"opendi/model-hub/api/testutils"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// TestMain is the entry point for the test suite. It sets up the environment and runs all tests.
func TestMain(m *testing.M) {
	// Setup code here (e.g., database connection, environment variables)
	setup()

	// Run tests
	code := m.Run()
	//by default, go runs tests sequentilaly IN THE SAME PACKAGE

	// Teardown code here (cleanup)
	teardown()

	// Exit with test result code
	os.Exit(code)
}

// setup initializes the database and loads environment variables
func setup() {

	//import environment variables
	err := godotenv.Load("../config/.env.test")
	if err != nil {
		fmt.Println("Error importing environment variables: ", err)
		os.Exit(1)
	}
	ret := 0

	//initialize DB instance
	ret, err = InitializeDBInstance()
	if ret != 0 {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}

	//reset the tables to have no contents.
	ResetTables()

}

// teardown cleans up resources after tests are run
func teardown() {
	// Clean up resources
}

func TestGetModelByUUID(t *testing.T) {
	ResetTables()

	t.Log("Running TestGetModelByUUID")
	CreateExampleData()

	//gets all models in the database
	models, _ := GetAllModels()

	if len(models) != 3 {
		t.Errorf("Expected 3 model, got %d", len(models))

	}

	//get the first model in the database
	model, err := GetModelByUUID(models[0].Meta.UUID)

	if err != nil {
		t.Errorf("Expected to not find model, err: %s", err)
	}

	if len(model.Meta.UUID) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(model.Meta.UUID))
	}

	//not the UUID
	anotherUUID := model.Meta.UUID + "1"

	_, err = GetModelByUUID(anotherUUID)

	if err == nil {
		t.Errorf("Expected to not find model, err: %s", err)
	}

}

func TestGetModelByTag(t *testing.T) {
	ResetTables()
	CreateExampleData()

	model, err := GetModelByTag("test-model:1.0")

	assert.NoError(t, err)
	assert.Equal(t, 36, len(model.Meta.UUID))
	assert.Equal(t, "Test Model", model.Meta.Name)
}

func TestCreateModel(t *testing.T) {
	ResetTables()
	CreateExampleData()

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
		MetaID: meta.ID,
		Meta:   meta,
	}

	resultModel, status, err := CreateModel(&model, 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, resultModel)

	// uuid and tag should have been generated
	assert.NotEmpty(t, resultModel.Meta.UUID)
	assert.Equal(t, "new-model:1.0", resultModel.Addons.Tag)

	assert.False(t, resultModel.Addons.IsPublic)
	assert.Nil(t, resultModel.Addons.Shares)
	assert.False(t, resultModel.Meta.CreatedAt.IsZero())
	assert.False(t, resultModel.Meta.UpdatedAt.IsZero())

	// should receive the same model when retrieving
	retrievedModel, err := GetModelByUUID(resultModel.Meta.UUID)
	assert.NoError(t, err)
	assert.Equal(t, resultModel.Meta.UUID, retrievedModel.Meta.UUID)
	assert.Equal(t, "new-model:1.0", retrievedModel.Addons.Tag)

	// attempting to create the same model again should report an error
	resultModel2, status, err := CreateModel(&model, 1)
	assert.Error(t, err)
	assert.Equal(t, http.StatusConflict, status)
	assert.Nil(t, resultModel2)
}

func TestUpdateModelAndCreateCommit(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// get the existing model
	oldModel, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)

	// should be no commits currently
	totalCommits, err := GetCommitsByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, totalCommits)

	// make changes to the model
	updatedModel := *oldModel
	updatedModel.Meta.Version = "2.0"
	updatedModel.Meta.Summary = "Updated test model"

	// update the model as user 1
	resultModel, status, err := UpdateModelAndCreateCommit(&updatedModel, oldModel, 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, resultModel)

	// verify model was updated
	assert.Equal(t, "2.0", resultModel.Meta.Version)
	assert.Equal(t, "Updated test model", resultModel.Meta.Summary)
	assert.Equal(t, "test-model:2.0", resultModel.Addons.Tag)

	// commit should have been created
	commit1, err := GetLatestCommitForModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, "2.0", commit1.Version)
	assert.Equal(t, -1, commit1.ParentID)
	assert.NotEmpty(t, commit1.Diff)

	// should be 1 total commit
	totalCommits, err = GetCommitsByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(totalCommits))

	// make changes to the model again
	updatedModel = *resultModel
	updatedModel.Meta.Version = "3.0"
	updatedModel.Meta.Summary = "Updated again"

	// update the model as user 1
	resultModel, status, err = UpdateModelAndCreateCommit(&updatedModel, resultModel, 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, resultModel)

	// verify model was updated
	assert.Equal(t, "3.0", resultModel.Meta.Version)
	assert.Equal(t, "Updated again", resultModel.Meta.Summary)
	assert.Equal(t, "test-model:3.0", resultModel.Addons.Tag)

	// commit properties should be set properly
	commit2, err := GetLatestCommitForModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, "3.0", commit2.Version)
	assert.Equal(t, commit1.ID, commit2.ParentID)
	assert.NotEmpty(t, commit2.Diff)

	// should now be 2 total commits
	totalCommits, err = GetCommitsByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(totalCommits))
}

func TestUpdateModelAndCreateCommitUnchangedVersion(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// get the existing model
	oldModel, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)

	// make changes to the model without changing the version
	updatedModel := *oldModel
	updatedModel.Meta.Summary = "Updated test model"

	// attempt to update the model as user 1
	resultModel, status, err := UpdateModelAndCreateCommit(&updatedModel, oldModel, 1)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, resultModel)

	// get the same model again
	oldModel, err = GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)

	// verify model was not updated
	assert.Equal(t, "1.0", oldModel.Meta.Version)
	assert.Equal(t, "This is a test model", oldModel.Meta.Summary)
	assert.Equal(t, "test-model:1.0", oldModel.Addons.Tag)

	// commit should not have been created
	commit, err := GetLatestCommitForModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, commit)
}

func TestUpdateModelAndCreateCommitUnchanged(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// get the existing model
	oldModel, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)

	// attempt to update the model as user 1 without making any changes
	resultModel, status, err := UpdateModelAndCreateCommit(oldModel, oldModel, 1)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, resultModel)

	// commit should not have been created
	commit, err := GetLatestCommitForModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, commit)
}

func TestGetUserByID(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// get a user that exists
	user, err := GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "creator@gmail.com", user.Email)

	// get a user that does not exist
	user, err = GetUserByID(5)
	assert.Error(t, err)
	assert.Nil(t, user)
}

// tests getting all models in the database
func TestGetAllModels(t *testing.T) {

	ResetTables()

	CreateExampleData()

	models, err := GetAllModels()
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" && models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}
	if models[1].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" && models[1].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}
}

func TestGetModelLineage(t *testing.T) {
	ResetTables()
	//example model is a parent-child pair.
	CreateExampleData()

	models, err := GetModelLineage("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e")
	if err != nil {
		t.Errorf("Expected to find model lineage, err: %s", err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 parent model, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		t.Errorf("Model doesn't match expected UUID")
	}
}

func TestGetModelChildren(t *testing.T) {
	ResetTables()
	//example model is a parent-child pair.
	CreateExampleData()

	models, err := GetModelChildren("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	if err != nil {
		t.Errorf("Expected to find model children, err: %s", err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 child model, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}

}

func TestIinitializingDbInstance(t *testing.T) {
	// Test that the environment variables are not set up
	// This test should fail if the environment variables are set up
	// This is because the environment variables are not necessary for the program to run

	username, _ := os.LookupEnv("OPEN_DI_DB_USERNAME")

	password, _ := os.LookupEnv("OPEN_DI_DB_PASSWORD")

	hostname, _ := os.LookupEnv("OPEN_DI_DB_HOSTNAME")

	port, _ := os.LookupEnv("OPEN_DI_DB_PORT")

	dbname, _ := os.LookupEnv("OPEN_DI_DB_NAME")

	os.Setenv("OPEN_DI_DB_USERNAME", "")
	_, err := InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}
	os.Setenv("OPEN_DI_DB_USERNAME", username)

	os.Setenv("OPEN_DI_DB_PASSWORD", password)
	os.Setenv("OPEN_DI_DB_HOSTNAME", "")

	_, err = InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}

	os.Setenv("OPEN_DI_DB_HOSTNAME", hostname)
	os.Setenv("OPEN_DI_DB_PORT", "")

	_, err = InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}

	os.Setenv("OPEN_DI_DB_PORT", port)
	os.Setenv("OPEN_DI_DB_NAME", "")

	_, err = InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}

	os.Setenv("OPEN_DI_DB_NAME", dbname)

	//_ = godotenv.Load("../config/.env.test") //note that godotenv doesn't set environment variables already set
	_, err = InitializeDBInstance()
	if err != nil {
		t.Errorf("Expected successful database initialization, got %s", err)
	}

	//tests giving database bad DSN
	os.Setenv("OPEN_DI_DB_USERNAME", "hahahaha")
	_, err = InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}

	//resets singleton variable
	os.Setenv("OPEN_DI_DB_USERNAME", username)
	_, err = InitializeDBInstance()
	if err != nil {
		t.Errorf("Expected successful database initialization, got %s", err)
	}

}

func TestCreateUser(t *testing.T) {
	ResetTables()

	user1, err1 := CreateUser(apiTypes.User{
		Username: "user1",
		Email:    "user1@example.com",
		GoogleID: "user1-googleid",
	})

	if err1 != nil {
		print(err1.Error())
	}

	if user1.Username != "user1" || user1.Email != "user1@example.com" || user1.GoogleID != "user1-googleid" {
		t.Fatalf("Username or password is not set correctly")
	}

	user1_copy, err1_1 := GetUserByEmail(user1.Email)
	if err1_1 != nil {
		t.Fatalf("Error when looking up user by email for user1")
	}

	if !user1.Equals(*user1_copy) {
		t.Fatalf("No error was thrown when getting user1 by email, but the user retrieved does not match the one created.")
	}

	//Now check that we can create more users without conflict
	user2, err2 := CreateUser(apiTypes.User{
		Username: "user2",
		Email:    "user2@example.com",
		GoogleID: "user2-googleid",
	})
	if err2 != nil {
		print(err2.Error())
	}

	if user2.Username != "user2" || user2.Email != "user2@example.com" || user2.GoogleID != "user2-googleid" {
		t.Fatalf("Username or password is not set correctly")
	}

	user2_copy, err2_2 := GetUserByEmail(user2.Email)
	if err2_2 != nil {
		t.Fatalf("Error when looking up user by email for user1")
	}

	if !user2.Equals(*user2_copy) {
		t.Fatalf("No error was thrown when getting user1 by email, but the user retrieved does not match the one created.")
	}

	//Ensure IDs are NOT equal
	if user1.ID == user2.ID {
		t.Fatalf("User 1's ID is the same as User 2's - this is extremely unlikely and almost certainly due to a bug.")
	}

	//Ensure we haven't regressed with User 1
	user1_copy, err1_1 = GetUserByEmail(user1.Email)
	if err1_1 != nil {
		t.Fatalf("Error when looking up user by email for user1")
	}

	if !user1.Equals(*user1_copy) {
		t.Fatalf("No error was thrown when getting user1 by email, but the user retrieved does not match the one created.")
	}
}

// also tests applyInvertedPatch
func TestGetAllCommits(t *testing.T) {
	ResetTables()
	CreateExampleData()
	ret, commits, error := GetAllCommits()
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(commits) != 0 {
		t.Errorf("Expected 0 commits, got %d", len(commits))
	}

	// create a commit
	models, err := GetAllModels()
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	expectedModel := models[0]
	if expectedModel.Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		expectedModel = models[1]
	}

	// Create a commit
	expectedModel.Meta.Summary = "changed!"
	expectedModel.Meta.Version = "2.0"

	oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	changedModel, status, err := UpdateModelAndCreateCommit(&expectedModel, oldModel, 1)

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	// Get all commits  There should be a new model created after updating the model.
	ret, commits, error = GetAllCommits()
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}
	if commits[0].ParentID != -1 {
		t.Errorf("Expected parent commit ID to be -1, got %d", commits[0].ParentID)
	}

	//try applying diff to get first model.
	//first get byte form of new model
	changedMdelBytes, _ := json.Marshal(changedModel)

	patchAppliedModel, err := jsonDiffHelpers.ApplyInvertedPatch(changedMdelBytes, []byte(commits[0].Diff))
	if err != nil {
		t.Errorf("Error applying patch: %s", err)
	}
	//get the old model bytes to compare to.
	oldModelBytes, _ := json.Marshal(oldModel)

	//reformat patchAppliedModel so that the JSON is in the correct order, not alphabetical. When JSON.marshal is claled on a certain struct type, fields will always be i n the same order.
	var tempModel *apiTypes.CausalDecisionModel
	json.Unmarshal(patchAppliedModel, &tempModel)
	patchAppliedModelBytes2, _ := json.Marshal(tempModel)

	if string(patchAppliedModelBytes2) != string(oldModelBytes) {
		t.Errorf("Expected model bytes to be equal, got %s and %s", string(patchAppliedModelBytes2), string(oldModelBytes))
	}

}

// testing create user given object given the same ID, which should throw an ID.
func TestCreateUserNonUniqueID(t *testing.T) {
	ResetTables()
	CreateExampleData() //also creates sample users
	user := apiTypes.User{
		ID: 1,
	}
	_, err := CreateUser(user)
	if err == nil {
		t.Errorf("Error should have been created when creating user")
	}

}

func TestFindOrCreateUserFromGoogleExisting(t *testing.T) {
	ResetTables()
	CreateExampleData()

	user, err := FindOrCreateUserFromGoogle(
		"creator",
		"creator@gmail.com",
		"creator-googleid",
		"",
	)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "creator", user.Username)
	assert.Equal(t, "creator@gmail.com", user.Email)
	assert.Equal(t, "creator-googleid", user.GoogleID)
}

func TestFindOrCreateUserFromGoogleNew(t *testing.T) {
	ResetTables()
	CreateExampleData()

	user, err := FindOrCreateUserFromGoogle(
		"newuser",
		"newuser@gmail.com",
		"newuser-googleid",
		"",
	)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "newuser@gmail.com", user.Email)
	assert.Equal(t, "newuser-googleid", user.GoogleID)

	// make sure it was created in the database
	var dbUser apiTypes.User
	err = dbInstance.Where("google_id = ?", "newuser-googleid").First(&dbUser).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, dbUser.ID)
}

// doesn't test that every single ID with corresopnding UUID has been matched yet.
func TestMatchUUIDToID(t *testing.T) {
	ResetTables()
	CreateExampleData()
	var model4 apiTypes.CausalDecisionModel
	err := testutils.LoadJSONFromFile("../test_files/model4.json", &model4)
	if err != nil {
		t.Fatalf("Error loading JSON file: %s", err)
	}
	model4.Diagrams[0].Dependencies[0].ID = 0
	model4.Diagrams[0].Dependencies[0].Meta.Creator.ID = 0
	model4.Diagrams[0].Dependencies[0].Meta.ID = 0
	model4.Diagrams[0].Elements[0].ID = 0
	model4.ID = 0

	transaction := dbInstance.Begin()
	if err := matchUUIDsToID(transaction, model4); err != nil {
		transaction.Rollback()
		t.Errorf("Error matching UUIDs to ID: %s", err)
	}
	_, _, err = CreateModel(&model4, 1)
	if err != nil {
		t.Errorf("Error creating model: %s", err)
	}
	transaction.Commit()
	// Check if the IDs are set correctly
	if model4.Diagrams[0].Dependencies[0].ID == 0 {
		t.Errorf("Error: ID should not be 0")
	}
	if model4.Diagrams[0].Dependencies[0].Meta.Creator.ID == 0 {

		t.Errorf("Error: Creator ID should not be 0")
	}
	if model4.Diagrams[0].Dependencies[0].Meta.ID == 0 {
		t.Errorf("Error: Meta ID should not be 0")
	}
	if model4.Diagrams[0].Elements[0].ID == 0 {
		t.Errorf("Error: Element ID should not be 0")
	}

}

// tests getting commit by ID
func TestGetCommitById(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// create a commit
	models, err := GetAllModels()
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	expectedModel := models[0]
	if expectedModel.Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		expectedModel = models[1]
	}

	// Create a commit
	expectedModel.Meta.Summary = "changed!"
	expectedModel.Meta.Version = "2.0"

	oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	// grfreema NOTE: check err state here
	_, _, _ = UpdateModelAndCreateCommit(&expectedModel, oldModel, 1)

	// grfreema NOTE: check err state here
	//get the commit
	_, commits, _ := GetAllCommits()

	commit := commits[0]

	status, commit2, err := GetCommitByID(commit.ID)
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	if commit.Diff != commit2.Diff {
		t.Errorf("Expected commit diff to be equal, got %s and %s", commit.Diff, commit2.Diff)
	}

	status, _, err = GetCommitByID(17)
	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusNotFound, status, err)
	}

}

func TestSearchModelsByName(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// Search for models by name
	models, err := SearchModelsByName("Child")
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

}

func TestSearchModelsByUser(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// Search for models by name
	models, err := SearchModelsByUser("Child")
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

}

func TestUpdateModelPrivacyByUUID(t *testing.T) {
	ResetTables()
	CreateExampleData()

	shares := []apiTypes.Share{
		{
			Email: "childcreator@gmail.com",
			Level: "write",
		},
	}

	err := UpdateModelPrivacyByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", true, shares)
	assert.NoError(t, err)

	// make sure the model was updated
	model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, true, model.Addons.IsPublic)
	assert.Equal(t, 1, len(model.Addons.Shares))
	assert.Equal(t, "childcreator@gmail.com", model.Addons.Shares[0].Email)
	assert.Equal(t, "write", model.Addons.Shares[0].Level)
}

func TestUpdateModelPrivacyByUUIDBadUUID(t *testing.T) {
	ResetTables()
	CreateExampleData()

	shares := []apiTypes.Share{}

	err := UpdateModelPrivacyByUUID("baduuid", false, shares)
	assert.Error(t, err)
}

func TestCreateAndGetTransfer(t *testing.T) {
	ResetTables()
	CreateExampleData()

	transfer := &apiTypes.Transfer{
		CDMUUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		FromUserID: 1,
		ToUserID:   2,
		CreatedAt:  time.Now(),
	}

	// create transfer
	err := CreateTransfer(transfer)
	assert.NoError(t, err)

	// make sure the transfer was created
	retrievedTransfer, err := GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, transfer.CDMUUID, retrievedTransfer.CDMUUID)
	assert.Equal(t, transfer.FromUserID, retrievedTransfer.FromUserID)
	assert.Equal(t, transfer.ToUserID, retrievedTransfer.ToUserID)
}

func TestGetTransferByModelUUIDNotFound(t *testing.T) {
	ResetTables()
	CreateExampleData()

	transfer, err := GetTransferByModelUUID("baduuid")
	assert.Error(t, err)
	assert.Nil(t, transfer)
}

func TestDeleteTransferAccept(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// create a transfer
	transfer := &apiTypes.Transfer{
		CDMUUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		FromUserID: 1,
		ToUserID:   2,
		CreatedAt:  time.Now(),
	}
	err := CreateTransfer(transfer)
	assert.NoError(t, err)

	// accept the transfer
	err = DeleteTransfer(transfer, true)
	assert.NoError(t, err)

	// ownership should have changed
	model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 2, model.Addons.OwnerID)

	// transfer should have been deleted
	deletedTransfer, err := GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, deletedTransfer)
}

func TestDeleteTransferDecline(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// create a transfer
	transfer := &apiTypes.Transfer{
		CDMUUID:    "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		FromUserID: 1,
		ToUserID:   2,
		CreatedAt:  time.Now(),
	}
	err := CreateTransfer(transfer)
	assert.NoError(t, err)

	// decline the transfer
	err = DeleteTransfer(transfer, false)
	assert.NoError(t, err)

	// ownership should not have changed
	model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 1, model.Addons.OwnerID)

	// transfer should have been deleted
	deletedTransfer, err := GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, deletedTransfer)
}
