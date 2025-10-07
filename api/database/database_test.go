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
	_, models, _ := GetAllModels()

	if len(models) != 2 {
		t.Errorf("Expected 2 model, got %d", len(models))

	}

	//get the first model in the database
	status, model, err := GetModelByUUID(models[0].Meta.UUID)

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	if len(model.Meta.UUID) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(model.Meta.UUID))
	}

	//not the UUID
	anotherUUID := model.Meta.UUID + "1"

	status, _, err = GetModelByUUID(anotherUUID)

	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusNotFound, status, err)
	}

}

// TestCreateModel tests the CreateModel function
func TestCreateModel(t *testing.T) {

	ResetTables()

	CreateExampleData()

	// There should be a user with id 2. Retrieve it.
	_, user, _ := GetUserByID(1)

	// Ensure the user is not nil
	if user == nil {
		t.Fatalf("User with ID 1 not found.")
	}

	meta := apiTypes.Meta{
		ID:            30,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6f",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     1,
		Creator:       *user,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{},
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        1234567890,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    meta.ID,
		Meta:      meta,
		Diagrams:  nil,
	}

	status, err := CreateModel(&model)
	if status != http.StatusCreated {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusCreated, status, err)
	}

	var model2 *apiTypes.CausalDecisionModel

	status, model2, err = GetModelByUUID(model.Meta.UUID)

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	if model.Meta.UUID != model2.Meta.UUID {
		t.Errorf("Models have differing UUID.")
	}

}

// tests getting all models in the database
func TestGetAllModels(t *testing.T) {

	ResetTables()

	CreateExampleData()

	ret, models, error := GetAllModels()
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" && models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}
	if models[1].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" && models[1].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}
}

// TestGetModelLineage tests the GetModelLineage function
func TestGetModelLineage(t *testing.T) {
	ResetTables()
	//example model is a parent-child pair.
	CreateExampleData()

	ret, models, error := GetModelLineage("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e")
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 parent model, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		t.Errorf("Model doesn't match expected UUID")
	}
}

// TestGetModelChildren tests the GetModelChildren function
// This function is used to get the children of a model given its UUID
func TestGetModelChildren(t *testing.T) {
	ResetTables()
	//example model is a parent-child pair.
	CreateExampleData()

	ret, models, error := GetModelChildren("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 child model, got %d", len(models))
	}

	if models[0].Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e" {
		t.Errorf("Model doesn't match expected UUID")
	}

}

// TestInitializingDBInstance tests the InitializeDBInstance function.
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

func TestCreateModelGivenEmail(t *testing.T) {
	ResetTables()

	//We need to create the user before we run the test
	creator, err := CreateUser(apiTypes.User{
		Username: "test",
		Email:    "test@example.com",
		GoogleID: "test-googleid",
	})

	// Ensure the user is not nil
	if err != nil {
		t.Fatalf("Unable to create test user.")
	}

	meta := apiTypes.Meta{
		ID:            30,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Name:          "Email Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		Creator:       *creator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{},
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        12367890,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    meta.ID,
		Meta:      meta,
		Diagrams:  nil,
	}

	//note that:
	//model.Meta gets a COPY of the previous meta object, meaning they are two separate Meta instances in memory.

	status, err := CreateModel(&model)

	if status != http.StatusCreated {
		t.Fatalf("There was an error when creating the model given the email. Status: %d Error:%s", status, err.Error())
	}

	status2, models, _ := GetAllModels()
	if status2 != http.StatusOK {
		t.Fatalf("Get all models failed.")
	}

	if models[0].Meta.Name != "Email Test Model" {
		t.Fatalf("The model was created successfully but the names do not match. \n Expected name: Email Test Model. \n Actual name: %s ", models[0].Meta.Name)

	}

	status, _, err = GetUserByEmail("nope")
	if status != http.StatusNotFound {
		t.Fatalf("There was an error when getting the user by email. Status: %d Error:%s", status, err.Error())
	}

	//tests creating a model with no user associated with the given email.

	/*
		meta.Creator.Email = "nope"
		fmt.Println("This was the email for the creator: ", model.Meta.Creator.Email)
		model.Meta.Creator.Email = "nope" //dont' forget that model.Meta is not the same underlying object as meta!
	*/
	model.Meta.Creator.Email = "nope" //dont' forget that model.Meta is not the same underlying object as Meta!
	//fmt.Println("This was the email for the creator: ", model.Meta.Creator.Email)

	// grfreema NOTE: check err state here
	status, err = CreateModel(&model)

	if status != http.StatusConflict {
		t.Fatalf("There was an error when creating the model given the email. Status: %d", status)
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

	status1, user1_copy, err1_1 := GetUserByEmail(user1.Email)
	if status1 != http.StatusOK || err1_1 != nil {
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

	status2, user2_copy, err2_2 := GetUserByEmail(user2.Email)
	if status2 != http.StatusOK || err2_2 != nil {
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
	status1, user1_copy, err1_1 = GetUserByEmail(user1.Email)
	if status1 != http.StatusOK || err1_1 != nil {
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
	status, models, err := GetAllModels()
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	expectedModel := models[0]
	if expectedModel.Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		expectedModel = models[1]
	}
	//prevSummary := expectedModel.Meta.Summary
	// Create a commit
	expectedModel.Meta.Summary = "changed!"

	// grfreema NOTE: check status state here
	status, oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	changedModel, status, err := UpdateModelAndCreateCommit(&expectedModel, oldModel)

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
	if commits[0].ParentCommitID != "" {
		t.Errorf("Expected parent commit ID to be empty, got %s", commits[0].ParentCommitID)
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
	_, err = CreateModel(&model4)
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
	status, models, err := GetAllModels()
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	expectedModel := models[0]
	if expectedModel.Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		expectedModel = models[1]
	}
	//prevSummary := expectedModel.Meta.Summary
	// Create a commit
	expectedModel.Meta.Summary = "changed!"

	_, oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	// grfreema NOTE: check err state here
	_, status, err = UpdateModelAndCreateCommit(&expectedModel, oldModel)

	// grfreema NOTE: check err state here
	//get the commit
	_, commits, err := GetAllCommits()

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

// also tests getting latest commit for model UUID
func TestUpdateModelAndCreateCommit(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// create a commit
	status, models, err := GetAllModels()
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	expectedModel := models[0]
	if expectedModel.Meta.UUID != "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" {
		expectedModel = models[1]
	}
	//prevSummary := expectedModel.Meta.Summary
	// Create a commit
	expectedModel.Meta.Summary = "changed!"

	_, oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	newmodel, status, err := UpdateModelAndCreateCommit(&expectedModel, oldModel)

	newmodelbytes, _ := json.Marshal(newmodel)
	expectedmodelbytes, _ := json.Marshal(expectedModel)
	if string(newmodelbytes) != string(expectedmodelbytes) {
		t.Errorf("Expected model to be equal, got %s and %s", newmodelbytes, expectedmodelbytes)
	}
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	//add another commit
	expectedModel.Meta.Summary = "changed again!"
	_, status, err = UpdateModelAndCreateCommit(newmodel, oldModel)
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	// grfreema NOTE: check status and err states
	// get latest commit
	status, commit, err := GetLatestCommitForModelUUID(expectedModel.Meta.UUID)

	//commit version should be 2, parent should not be ""
	if commit.Version != 2 {
		t.Errorf("Expected commit version 2, got %d", commit.Version)
	}
	if commit.ParentCommitID == "" {
		t.Errorf("Expected parent commit ID to be not empty, got %s", commit.ParentCommitID)
	}

}

func TestSearchModelsByName(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// Search for models by name
	status, models, err := SearchModelsByName("Child")
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

}

func TestSearchModelsByUser(t *testing.T) {
	ResetTables()
	CreateExampleData()

	// Search for models by name
	status, models, err := SearchModelsByUser("Child")
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
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
			UserID: 2,
			Level:  "write",
		},
	}

	err := UpdateModelPrivacyByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d", true, shares)
	assert.NoError(t, err)

	// make sure the model was updated
	_, model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, true, model.IsPublic)
	assert.Equal(t, 1, len(model.Shares))
	assert.Equal(t, 2, model.Shares[0].UserID)
	assert.Equal(t, "write", model.Shares[0].Level)
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
	_, model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 2, model.OwnerID)

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
	_, model, err := GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.NoError(t, err)
	assert.Equal(t, 1, model.OwnerID)

	// transfer should have been deleted
	deletedTransfer, err := GetTransferByModelUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	assert.Error(t, err)
	assert.Nil(t, deletedTransfer)
}
