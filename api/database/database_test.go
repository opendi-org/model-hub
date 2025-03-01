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
	//by default, go runs tests sequentilaly

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
	//we also test the initialize DB instance here
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
	t.Log("Running TestGetModelByUUID")
	CreateExampleModel()

	status, model, err := GetModelByUUID("1234-5678-9101")

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	if model.Meta.UUID != "1234-5678-9101" {
		t.Errorf("Expected model UUID %s, got %s", "1234-5678-9101", model.Meta.UUID)
	}

	status, _, err = GetModelByUUID("1234-5678-9103")

	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusNotFound, status, err)
	}

}

func TestCreateModel(t *testing.T) {

	// There should be a user with id 2. Retrieve it.
	_, user, _ := GetUserByID(2)

	// Ensure the user is not nil
	if user == nil {
		t.Fatalf("User with ID 2 not found.")
	}

	meta := apiTypes.Meta{
		ID:            30,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1234-5678-9105",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     1,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{*user},
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

	status, model2, err = GetModelByUUID("1234-5678-9105")

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, status, err)
	}

	if model.Meta.UUID != model2.Meta.UUID {
		t.Errorf("Models have differing UUID.")
	}

}

func TestGetAllModels(t *testing.T) {
	ret, models, error := GetAllModels()
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}

	if models[0].Meta.UUID != "1234-5678-9101" {
		t.Errorf("Expected model UUID %s, got %s", "1234-5678-9101", models[0].Meta.UUID)
	}
	if models[2].Meta.UUID != "1234-5678-9105" {
		t.Errorf("Expected model UUID %s, got %s", "1234-5678-9105", models[2].Meta.UUID)
	}

}
