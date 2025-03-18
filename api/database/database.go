//
// COPYRIGHT OpenDI
//

package database

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// global db instance
var dbInstance *gorm.DB

func CreateTablesIfNotCreated() error {

	// AutoMigrate all the structs defined in apitypes.go
	err := dbInstance.AutoMigrate(
		&apiTypes.CausalDecisionModel{},
		&apiTypes.User{},
		&apiTypes.Meta{},
		&apiTypes.Diagram{},
		&apiTypes.DiaElement{},
		&apiTypes.CausalDependency{},
	)
	return err

}

func ResetTables() {

	dbInstance := GetDBInstance()

	// Drop all tables
	var tables []string
	dbInstance.Raw("SHOW TABLES").Scan(&tables) // Get all table names

	for _, table := range tables {
		dbInstance.Migrator().DropTable(table)
	}

	CreateTablesIfNotCreated()

}

// initialize db instance
func InitializeDBInstance() (int, error) {

	// Construct the Data Source Name (DSN) for the database connection

	// Check to make sure the environment variables for the database connection are set before using them
	username, ok := os.LookupEnv("OPEN_DI_DB_USERNAME")
	if !ok || username == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_USERNAME is not set or empty")
	}
	password, ok := os.LookupEnv("OPEN_DI_DB_PASSWORD")
	if !ok || password == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_PASSWORD is not set or empty")
	}
	hostname, ok := os.LookupEnv("OPEN_DI_DB_HOSTNAME")
	if !ok || hostname == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_HOSTNAME is not set or empty")
	}
	port, ok := os.LookupEnv("OPEN_DI_DB_PORT")
	if !ok || port == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_PORT is not set or empty")
	}
	dbname, ok := os.LookupEnv("OPEN_DI_DB_NAME")
	if !ok || dbname == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_NAME is not set or empty")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, hostname, port, dbname)

	var err error
	if dbInstance != nil {
		sqlDB, _ := dbInstance.DB()
		sqlDB.Close()
		dbInstance = nil

	}

	dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		dbInstance = nil
		return 1, err
	}

	err = CreateTablesIfNotCreated()
	if err != nil {
		return 1, err
	}

	return 0, nil

}

// gets singleton db instance
func GetDBInstance() *gorm.DB {
	return dbInstance
}

// function for getting all models in Go struct  - remember, in Go, public methods have to be capitalized
func GetAllModels() (int, []apiTypes.CausalDecisionModel, error) {
	var models []apiTypes.CausalDecisionModel
	// Updated query to preload associated fields
	if err := dbInstance.
		Preload("Meta").
		Preload("Diagrams").
		Preload("Diagrams.Meta").
		Preload("Diagrams.Elements").
		Preload("Diagrams.Dependencies").
		Preload("Diagrams.Elements.Meta").
		Preload("Diagrams.Dependencies.Meta").
		Preload("Meta.Creator").
		Preload("Meta.Updaters").
		Find(&models).Error; err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, models, nil

}

// Example method that creates a sample model in the database
func CreateExampleModel() {
	creator := apiTypes.User{
		ID:       1,
		UUID:     "user-uuid-creator",
		Username: "Test Creator",
		Email:    "creator@example.com",
		Password: "p",
	}

	updater := apiTypes.User{
		ID:       2,
		UUID:     "user-uuid-updater",
		Username: "Test Updater",
		Email:    "updater@example.com",
		Password: "q",
	}

	meta := apiTypes.Meta{
		ID:            1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1234-5678-9101",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     creator.ID,
		Creator:       creator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{updater},
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    1,
		Meta:      meta,
		Parent:    nil,
		Diagrams:  nil,
	}

	if err := dbInstance.Create(&model).Error; err != nil {
		fmt.Println("Error creating model: ", err)
	}

	// Also create a child model
	childCreator := apiTypes.User{
		ID:       3,
		UUID:     "user-uuid-child-creator",
		Username: "Test Child Creator",
		Email:    "mail.com",
		Password: "p",
	}

	childUpdater := apiTypes.User{
		ID:       4,
		UUID:     "user-uuid-child-updater",
		Username: "Test Child Updater",
		Email:    "mail.com",
		Password: "q",
	}

	childMeta := apiTypes.Meta{
		ID:            2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1324-5678-9101",
		Name:          "Test Child Model",
		Summary:       "This is a test child model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     childCreator.ID,
		Creator:       childCreator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{childUpdater, updater},
		UpdatedDate:   "2021-07-01",
	}

	childModel := apiTypes.CausalDecisionModel{
		ID:         2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Schema:     "Test Child Schema",
		MetaID:     2,
		Meta:       childMeta,
		ParentUUID: model.Meta.UUID,
		ParentID:   &model.ID,
		Parent:     &model,
		Diagrams:   nil,
	}

	if err := dbInstance.Create(&childModel).Error; err != nil {
		fmt.Println("Error creating child model: ", err)
	}

}

// CreateModel encapsulates the GORM functionality for creating a model with its metadata in a transaction
func CreateModel(uploadedModel *apiTypes.CausalDecisionModel) (int, error) {
	// Ensure no other model with the same UUID exists.
	var count int64
	dbInstance.Model(&apiTypes.Meta{}).Where("uuid = ?", uploadedModel.Meta.UUID).Count(&count)
	if count > 0 {
		// If a meta with the same UUID exists, return a conflict error.
		return http.StatusConflict, fmt.Errorf("a model with UUID %s already exists", uploadedModel.Meta.UUID)
	}

	// Begin transaction.
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	// Try to retrieve creator id information from the meta, then find a creator with that id in the database.

	var creator apiTypes.User
	var countCreator int64
	transaction.Model(&apiTypes.User{}).Where("uuid = ?", uploadedModel.Meta.Creator.UUID).Count(&countCreator)
	if countCreator == 0 {
		// Create the creator in the database if it does not exist.
		if err := transaction.Create(&uploadedModel.Meta.Creator).Error; err != nil {
			transaction.Rollback()
			return http.StatusInternalServerError, fmt.Errorf("could not create creator: %s", err.Error())
		}
	} else {
		// Find the creator in the database using the uuid
		if err := transaction.Where("uuid = ?", uploadedModel.Meta.Creator.UUID).First(&creator).Error; err != nil {
			transaction.Rollback()
			return http.StatusInternalServerError, fmt.Errorf("could not find creator: %s", err.Error())
		}
		fmt.Println(creator)
		uploadedModel.Meta.Creator = creator
	}

	// Try to retrieve updater id information from the meta, then find an updater with that id in the database.
	for i, updater := range uploadedModel.Meta.Updaters {
		var countUpdater int64
		transaction.Model(&apiTypes.User{}).Where("uuid = ?", updater.UUID).Count(&countUpdater)
		if countUpdater == 0 {
			// Create the updater in the database if it does not exist.
			if err := transaction.Create(&uploadedModel.Meta.Updaters[i]).Error; err != nil {
				transaction.Rollback()
				return http.StatusInternalServerError, fmt.Errorf("could not create updater: %s", err.Error())
			}
		} else {
			// Find the updater in the database using the uuid
			if err := transaction.Where("uuid = ?", updater.UUID).First(&uploadedModel.Meta.Updaters[i]).Error; err != nil {
				transaction.Rollback()
				return http.StatusInternalServerError, fmt.Errorf("could not find updater: %s", err.Error())
			}
		}
	}

	// Create meta in transaction; error out on failure.
	if err := transaction.Create(&uploadedModel.Meta).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not create model meta: %s", err.Error())
	}

	// Create the model in transaction; error out on failure.
	if err := transaction.Create(&uploadedModel).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not create model: %s", err.Error())
	}

	// Commit the transaction; error out if commit fails.
	if err := transaction.Commit().Error; err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not commit transaction: %s", err.Error())
	}

	return http.StatusCreated, nil
}

// GetModelByUUID encapsulates the GORM functionality for getting a model by its UUID
func GetModelByUUID(uuid string) (int, *apiTypes.CausalDecisionModel, error) {
	var meta apiTypes.Meta

	// Find the meta record with the given UUID.
	if err := dbInstance.Where("uuid = ?", uuid).First(&meta).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("meta with uuid %s not found", uuid)
	}

	var model apiTypes.CausalDecisionModel

	// Find the model that has the found meta record, preloading associated fields.
	if err := dbInstance.
		Preload("Meta").
		Preload("Diagrams").
		Preload("Diagrams.Meta").
		Preload("Diagrams.Elements").
		Preload("Diagrams.Dependencies").
		Preload("Diagrams.Elements.Meta").
		Preload("Diagrams.Dependencies.Meta").
		Preload("Meta.Creator").
		Preload("Meta.Updaters").
		Where("meta_id = ?", meta.ID).
		First(&model).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("this meta is not associated with a model")
	}

	return http.StatusOK, &model, nil
}

// GetUserByID encapsulates the GORM functionality for getting a user by their ID
func GetUserByID(id int) (int, *apiTypes.User, error) {
	var user apiTypes.User

	// Find the user record with the given ID.
	if err := dbInstance.Where("id = ?", id).First(&user).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("user with id %d not found", id)
	}

	return http.StatusOK, &user, nil
}

func GetUserByEmail(email string) (int, *apiTypes.User, error) {
	var user apiTypes.User

	// Find the user record with the given ID.
	if err := dbInstance.Where("email = ?", email).First(&user).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("user with email %s not found", email)
	}

	return http.StatusOK, &user, nil
}

func CreateUser(email string, password string) (*apiTypes.User, error) {
	var newuser apiTypes.User
	newuser.Username = email
	newuser.Email = email
	newuser.Password = password
	newuser.UUID = uuid.NewString()
	// Ensure no other user with this email exists
	var count int64
	dbInstance.Model(&apiTypes.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		// If a meta with the same UUID exists, return a conflict error.
		return nil, fmt.Errorf("a user with email %s already exists", email)
	}

	// Begin transaction.
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return nil, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	if err := transaction.Create(&newuser).Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("could not create updater: %s", err.Error())
	}

	transaction.Commit()

	return &newuser, nil
}

func UserLogin(email string, password string) (int, *apiTypes.User, error) {

	status, user, _ := GetUserByEmail(email)

	if status != 200 {
		//For now, let's just create a new user
		newuser, err := CreateUser(email, password)
		if err != nil {
			return http.StatusConflict, nil, fmt.Errorf("user does not exist and could not create new user")
		}
		return http.StatusOK, newuser, nil
	}

	return http.StatusOK, user, nil
}

func GetModelLineage(uuid string) (int, []apiTypes.CausalDecisionModel, error) {
	status, modelPtr, err := GetModelByUUID(uuid)

	if err != nil {
		return status, nil, err
	}

	model := *modelPtr

	var lineage []apiTypes.CausalDecisionModel

	for model.ParentUUID != "" {
		_, parentPtr, err := GetModelByUUID(model.ParentUUID)

		if err != nil {
			break
		}

		parent := *parentPtr
		lineage = append(lineage, parent)
		model = parent
	}

	// Reverse the lineage so that the earliest ancestor is first.
	for i, j := 0, len(lineage)-1; i < j; i, j = i+1, j-1 {
		lineage[i], lineage[j] = lineage[j], lineage[i]
	}

	return http.StatusOK, lineage, nil
}
