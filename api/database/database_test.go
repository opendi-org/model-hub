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

	status, model, err = GetModelByUUID("1234-5678-9103")

	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusNotFound, status, err)
	}

}

func TestCreateModel(t *testing.T) {
	meta := apiTypes.Meta{
		ID:            2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1234-5678-9105",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		Creator:       "Test Creator",
		CreatedDate:   "2021-07-01",
		Updator:       "Test Updator",
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    2,
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

	if !model.Equals(*model2) {
		t.Errorf("Models are not equal")
	}

}

func TestGetAllModels(t *testing.T) {
	ret, models, error := GetAllModels()
	if ret != http.StatusOK {
		t.Errorf("Expected status %d, got %d, err: %s", http.StatusOK, ret, error)
	}
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	if models[0].Meta.UUID != "1234-5678-9101" {
		t.Errorf("Expected model UUID %s, got %s", "1234-5678-9101", models[0].Meta.UUID)
	}
	if models[1].Meta.UUID != "1234-5678-9105" {
		t.Errorf("Expected model UUID %s, got %s", "1234-5678-9105", models[1].Meta.UUID)
	}

}
