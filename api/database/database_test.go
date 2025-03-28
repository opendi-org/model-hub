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
)

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

func setup() {

	//import environment variables
	err := godotenv.Load("../config/.env.test")
	if err != nil {
		fmt.Println("Error importing environment variables: ", err)
		os.Exit(1)
	}
	ret := 0

	ret, err = InitializeDBInstance()
	if ret != 0 {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}

	ResetTables()

}

func teardown() {
	// Clean up resources
}

func TestGetModelByUUID(t *testing.T) {
	ResetTables()

	t.Log("Running TestGetModelByUUID")
	CreateExampleModels()

	//gets first user in database
	//_, user, _ := GetUserByID(1)

	//gets all models in the database
	_, models, _ := GetAllModels()

	if len(models) != 2 {
		t.Errorf("Expected 2 model, got %d", len(models))

	}

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

func TestCreateModel(t *testing.T) {

	ResetTables()

	CreateExampleModels()

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

	status, model2, err = GetModelByUUID("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6f")

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	if model.Meta.UUID != model2.Meta.UUID {
		t.Errorf("Models have differing UUID.")
	}

	//tests creating a model with the same UUID

}

func TestGetAllModels(t *testing.T) {

	ResetTables()

	CreateExampleModels()

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

func TestGetModelLineage(t *testing.T) {
	ResetTables()
	CreateExampleModels()

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

func TestGetModelChildren(t *testing.T) {
	ResetTables()
	CreateExampleModels()

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

func TestIinitializingDbInstance(t *testing.T) {
	// Test that the environment variables are not set up
	// This test should fail if the environment variables are set up
	// This is because the environment variables are not necessary for the program to run
	// They are only necessary for the program to run in a specific environment

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
	os.Setenv("OPEN_DI_DB_PASSWORD", "")

	_, err = InitializeDBInstance()
	if err == nil {
		t.Errorf("Expected error initializing database, got nil")
	}

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

	//tests whether this reset database instance if currently not nil
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
	creator, err := CreateUser("testgivenemail", "pass")

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
	//model.Meta gets a COPY of meta, meaning they are two separate Meta instances in memory.

	status, err := CreateModelGivenEmail(&model)

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
}

func TestCreateUser(t *testing.T) {
	//IMPORTANT NOTE: In the current implementation, the user's email and username are the same!
	//If/when this is eventually changed, this test must be edited! For now tests are written on the assumption
	//that email and username are the same.
	ResetTables()

	user1, err1 := CreateUser("user1", "pass1")

	if err1 != nil {
		print(err1.Error())
	}

	if user1.Email != "user1" || user1.Username != "user1" || user1.Password != "pass1" {
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
	user2, err2 := CreateUser("user2", "pass2")
	if err2 != nil {
		print(err2.Error())
	}

	if user2.Email != "user2" || user2.Username != "user2" || user2.Password != "pass2" {
		t.Fatalf("Username or password is not set correctly")
	}

	status2, user2_copy, err2_2 := GetUserByEmail(user2.Email)
	if status2 != http.StatusOK || err2_2 != nil {
		t.Fatalf("Error when looking up user by email for user1")
	}

	if !user2.Equals(*user2_copy) {
		t.Fatalf("No error was thrown when getting user1 by email, but the user retrieved does not match the one created.")
	}

	//Ensure UUIDs are NOT equal
	if user1.UUID == user2.UUID {
		t.Fatalf("User 1's UUID is the same as User 2's - this is extremely unlikely and almost certainly due to a bug.")
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

func TestUserLogin(t *testing.T) {
	//As with TestCreateUser, it is important to note that this test was written with the assumption that
	//the username and email are the same. If/when this is changed, make sure to edit this test!
	ResetTables()

	//Let's first login with a user that has not been created yet, and check that the user is properly created
	status1, user1, err1 := UserLogin("email1", "pass1")
	if status1 != http.StatusOK || err1 != nil {
		t.Fatalf("Error was thrown when trying to login a brand new user")
	}

	//Now let's check that the user was actually created
	status1_1, user1_copy, err1_1 := GetUserByEmail("email1")
	if status1_1 != http.StatusOK || err1_1 != nil {
		t.Fatalf("Error when trying to retrieve new user: %s", err1_1.Error())
	}

	if user1_copy.UUID != user1.UUID {
		t.Fatalf("UUID's do not match between user object retrieved upon login, and user lookup by email")
	}

	//Now we can try and login again, but with a wrong email
	status2, _, _ := UserLogin("email1", "wrong_password")
	if status2 == http.StatusConflict {
		t.Fatal("Trying to login with the wrong password throws an error that the user does not exist or there was some kind of database conflict.")
	} else if status2 != http.StatusUnauthorized {
		t.Fatal("User was able to login with the wrong password.")

	}

	status, _, err = GetUserByEmail("nope")
	if status != http.StatusNotFound {
		t.Fatalf("There was an error when getting the user by email. Status: %d Error:%s", status, err.Error())
	}

	//tests creating a model with an incorrect email for the associated user.

	/*
		meta.Creator.Email = "nope"
		fmt.Println("This was the email for the creator: ", model.Meta.Creator.Email)
		model.Meta.Creator.Email = "nope" //dont' forget that model.Meta is not the same underlying object as Meta!
	*/
	model.Meta.Creator.Email = "nope" //dont' forget that model.Meta is not the same underlying object as Meta!
	//fmt.Println("This was the email for the creator: ", model.Meta.Creator.Email)

	status, err = CreateModelGivenEmail(&model)

	if status != http.StatusConflict {
		t.Fatalf("There was an error when creating the model given the email. Status: %d", status)
	}

}

// also test applyInvertedPatch
func TestGetAllCommits(t *testing.T) {
	ResetTables()
	CreateExampleModels()
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

	status, oldModel, _ := GetModelByUUID(expectedModel.Meta.UUID)

	changedModel, status, err := UpdateModelAndCreateCommit(&expectedModel, oldModel)

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	// Get all commits
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

	//reformat patchAppliedModel so that the JSON is in the correct order, not alphabetical
	var tempModel *apiTypes.CausalDecisionModel
	json.Unmarshal(patchAppliedModel, &tempModel)
	patchAppliedModelBytes2, _ := json.Marshal(tempModel)

	if string(patchAppliedModelBytes2) != string(oldModelBytes) {
		t.Errorf("Expected model bytes to be equal, got %s and %s", string(patchAppliedModelBytes2), string(oldModelBytes))
	}

}

func TestCreateUserGivenObject(t *testing.T) {
	ResetTables()
	CreateExampleModels() //also creates sample users
	user := apiTypes.User{
		ID: 1,
	}
	_, err := createUserGivenObject(user)
	if err == nil {
		t.Errorf("Error should have been created when creating user")
	}

}

// doesn't test that every single ID with corresopnding UUID has been matched yet.
// TODO note to Isaac - for consistency, should every UUID field be stored not in the Meta object, but the actual object?
func TestMatchUUIDToID(t *testing.T) {
	ResetTables()
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

func TestGetCommitById(t *testing.T) {
	ResetTables()
	CreateExampleModels()

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

	_, status, err = UpdateModelAndCreateCommit(&expectedModel, oldModel)

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

func TestUpdateModelAndCreateCommit(t *testing.T) {
	ResetTables()
	CreateExampleModels()

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

func TestUserLogin(t *testing.T) {
	ResetTables()
	CreateExampleModels()

	//for now, failed login creates a user
	status, _, err := UserLogin("heehee", "password")
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	status, _, err = UserLogin("creator@example.com", "x")
	if status != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusUnauthorized, status, err)
	}

	status, _, err = UserLogin("creator@example.com", "p")
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusUnauthorized, status, err)
	}

}

func TestSearchModelsByName(t *testing.T) {
	ResetTables()
	CreateExampleModels()

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
	CreateExampleModels()

	// Search for models by name
	status, models, err := SearchModelsByUser("Child")
	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

}
