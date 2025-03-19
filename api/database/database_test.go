package database

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
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
