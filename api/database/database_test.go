package database

import (
	"os"
	"github.com/joho/godotenv"
	"testing"
	"fmt"
	"net/http"
)

func TestMain(m *testing.M) {
	// Setup code here (e.g., database connection, environment variables)
	setup()

	// Run tests
	code := m.Run()

	// Teardown code here (cleanup)
	teardown()

	// Exit with test result code
	os.Exit(code)
}

func resetTables() {

	dbInstance := GetDBInstance()

	// Drop all tables
	var tables []string
	dbInstance.Raw("SHOW TABLES").Scan(&tables) // Get all table names
	
	for _, table := range tables {
		dbInstance.Migrator().DropTable(table)
	}

	createTablesIfNotCreated()

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

	resetTables()

	
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