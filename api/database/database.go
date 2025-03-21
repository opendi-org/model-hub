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
		&apiTypes.Commit{},
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
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_USERNAME is not set or empty")
		os.Exit(1)
	}
	password, ok := os.LookupEnv("OPEN_DI_DB_PASSWORD")
	if !ok || password == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_PASSWORD is not set or empty")
		os.Exit(1)
	}
	hostname, ok := os.LookupEnv("OPEN_DI_DB_HOSTNAME")
	if !ok || hostname == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_HOSTNAME is not set or empty")
		os.Exit(1)
	}
	port, ok := os.LookupEnv("OPEN_DI_DB_PORT")
	if !ok || port == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_PORT is not set or empty")
		os.Exit(1)
	}
	dbname, ok := os.LookupEnv("OPEN_DI_DB_NAME")
	if !ok || dbname == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_NAME is not set or empty")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, hostname, port, dbname)

	var err error
	if dbInstance != nil {
		return 0, nil
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

// function for getting all commits in Go struct  - remember, in Go, public methods have to be capitalized
func GetAllCommits() (int, []apiTypes.Commit, error) {
	var commits []apiTypes.Commit
	// Updated query to preload associated fields
	if err := dbInstance.
		Find(&commits).Error; err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, commits, nil

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

// matchUUIDsToID recursively iterates through a CDM (or really any CDM component)
// and its nested structures and finds matching UUIDs in the database and updates
// the IDs of the components to match the ID found in the database
// It is designed to work with the structs defined in apitypes.go,
// which as of now are CausalDecisionModel, Meta, Diagram, DiaElement, CausalDependency, User,
// and Commit.
func matchUUIDsToID(tx *gorm.DB, component any) error {
	// Check if this is a Meta struct and create its users if they don't exist
	// While it may not make sense to have a meta get updated when performing a create, it is
	// necessary to do this for putting a model, since in that use case we are not creating a new meta,
	// but rather updating an existing one.
	// Furthermore, we should already be checking to make sure a meta with the same UUID does not exist in the database
	// before creating a new model, and so this should not be a problem. Meanwhile if we are creating a new
	// model and we are referencing, say for example, preexisting diagrams, we should be
	// getting the existing diagram and it's meta, not creating a new meta. So while it may see odd
	// to have code here that doesn't throw an error if the meta is found, it is necessary to not throw
	// and in fact makes sense to do so.
	if meta, ok := component.(*apiTypes.Meta); ok && meta.UUID != "" {
		var existingMeta apiTypes.Meta
		if err := tx.Where("uuid = ?", meta.UUID).First(&existingMeta).Error; err == nil {
			meta.ID = existingMeta.ID
			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if meta.CreatedAt.IsZero() {
				meta.CreatedAt = existingMeta.CreatedAt
			}
		}

		// TODO: Change this so we no longer create a new user if the UUID is not found
		// Right now this is just a workaround to create a new user, but in the future when
		// we have a way to properly create users, we should not do this, and instead if there
		// is no user with the UUID, we should error out and not create a new user.

		// Match Creator UUID to ID
		if meta.Creator.UUID != "" {
			var existingUser apiTypes.User
			if err := tx.Where("uuid = ?", meta.Creator.UUID).First(&existingUser).Error; err == nil {
				meta.Creator = existingUser
				meta.CreatorID = existingUser.ID
			} else if meta.Creator.ID == 0 {
				// Create user if not exists
				if err := tx.Create(&meta.Creator).Error; err != nil {
					return fmt.Errorf("could not create creator: %s", err.Error())
				}
				meta.CreatorID = meta.Creator.ID
			}
		}

		// Match Updaters UUIDs to IDs
		for i, updater := range meta.Updaters {
			if updater.UUID != "" {
				var existingUser apiTypes.User
				if err := tx.Where("uuid = ?", updater.UUID).First(&existingUser).Error; err == nil {
					meta.Updaters[i] = existingUser
				} else if updater.ID == 0 {
					// Create updater if not exists
					if err := tx.Create(&meta.Updaters[i]).Error; err != nil {
						return fmt.Errorf("could not create updater: %s", err.Error())
					}
				}
			}
		}
		return nil
	}

	// Check if this is a User struct and match its UUID to ID
	if user, ok := component.(*apiTypes.User); ok && user.UUID != "" {
		var existingUser apiTypes.User
		if err := tx.Where("uuid = ?", user.UUID).First(&existingUser).Error; err == nil {
			// Match the existing user ID to the current user
			user.ID = existingUser.ID
		}
		return nil
	}

	// Check if this is a CausalDecisionModel struct and recursively match its components' UUIDs to IDs
	if cdm, ok := component.(*apiTypes.CausalDecisionModel); ok {
		// Match Meta
		if err := matchUUIDsToID(tx, &cdm.Meta); err != nil {
			return err
		}

		// Try to find the existing CausalDecisionModel in the database
		var existingModel apiTypes.CausalDecisionModel

		// Check if the meta ID is set, if not, we should not try to find it in the database
		// since it is not a pre-existing model, but rather a new one
		if err := tx.Where("meta_id = ?", cdm.Meta.ID).First(&existingModel).Error; err == nil {
			cdm.ID = existingModel.ID

			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if cdm.CreatedAt.IsZero() {
				cdm.CreatedAt = existingModel.CreatedAt
			}
		}

		// Match Diagrams
		for i := range cdm.Diagrams {
			if err := matchUUIDsToID(tx, &cdm.Diagrams[i]); err != nil {
				return err
			}
		}

		// Match Parent if exists
		if cdm.ParentUUID != "" {
			var parentMeta apiTypes.Meta
			if err := tx.Where("uuid = ?", cdm.ParentUUID).First(&parentMeta).Error; err == nil {
				var parentModel apiTypes.CausalDecisionModel
				if err := tx.Where("meta_id = ?", parentMeta.ID).First(&parentModel).Error; err == nil {
					cdm.ParentID = &parentModel.ID
				}
			}
		}

		return nil
	}

	// Check if this is a Diagram struct and recursively match its components' UUIDs to IDs
	if diagram, ok := component.(*apiTypes.Diagram); ok {
		// Match Meta
		if err := matchUUIDsToID(tx, &diagram.Meta); err != nil {
			return err
		}

		// Try to find the existing Diagram in the database
		var existingDiagram apiTypes.Diagram

		// Check if the meta ID is set, if not, we should not try to find it in the database
		// since it is not a pre-existing diagram, but rather a new one
		if err := tx.Where("meta_id = ?", diagram.Meta.ID).First(&existingDiagram).Error; err == nil {
			diagram.ID = existingDiagram.ID

			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if diagram.CreatedAt.IsZero() {
				diagram.CreatedAt = existingDiagram.CreatedAt
			}
		}

		// Match Elements
		for i := range diagram.Elements {
			if err := matchUUIDsToID(tx, &diagram.Elements[i]); err != nil {
				return err
			}
		}

		// Match Dependencies
		for i := range diagram.Dependencies {
			if err := matchUUIDsToID(tx, &diagram.Dependencies[i]); err != nil {
				return err
			}
		}

		return nil
	}

	// Check if this is a DiaElement struct and match its Meta UUID to ID
	// then see if we can find the existing DiaElement in the database
	if element, ok := component.(*apiTypes.DiaElement); ok {
		// First match the meta UUID
		if err := matchUUIDsToID(tx, &element.Meta); err != nil {
			return err
		}
		// Try to find the existing DiaElement in the database
		var existingElement apiTypes.DiaElement

		// Check if the meta ID is set, if not, we should not try to find it in the database
		// since it is not a pre-existing element, but rather a new one
		if err := tx.Where("meta_id = ?", element.Meta.ID).First(&existingElement).Error; err == nil {
			element.ID = existingElement.ID

			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if element.CreatedAt.IsZero() {
				element.CreatedAt = existingElement.CreatedAt
			}
		}
		return nil
	}

	// Check if this is a CausalDependency struct and match its Meta UUID to ID
	// then see if we can find the existing CausalDependency in the database
	if dependency, ok := component.(*apiTypes.CausalDependency); ok {
		// First match the meta UUID
		if err := matchUUIDsToID(tx, &dependency.Meta); err != nil {
			return err
		}
		// Try to find the existing CausalDependency in the database
		var existingDependency apiTypes.CausalDependency

		// Check if the meta ID is set, if not, we should not try to find it in the database
		// since it is not a pre-existing dependency, but rather a new one
		if err := tx.Where("meta_id = ?", dependency.Meta.ID).First(&existingDependency).Error; err == nil {
			dependency.ID = existingDependency.ID

			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if dependency.CreatedAt.IsZero() {
				dependency.CreatedAt = existingDependency.CreatedAt
			}
		}
		return nil
	}

	// Check if this is a Commit struct and match its (supposedly unique) pair [ParentCommitID, CDMUUID] to ID
	if commit, ok := component.(*apiTypes.Commit); ok {
		// Try to find the existing Commit in the database
		var existingCommit apiTypes.Commit
		if err := tx.Where("parent_commit_id = ? AND cdm_uuid = ?", commit.ParentCommitID, commit.CDMUUID).First(&existingCommit).Error; err == nil {
			commit.ID = existingCommit.ID

			// Also if the created at time is zero, go ahead and set it to the existing created at time
			// This is necessary to fix a bug with PUT endpoints not sending a created at time thereby causing an invalid time to be set
			// which the database/GORM does not like
			if commit.CreatedAt.IsZero() {
				commit.CreatedAt = existingCommit.CreatedAt
			}
		}
	}

	return nil
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

	// Match all UUIDs in the model to existing database IDs where possible
	// This will ensure that we are not duplicating pre-existing components
	// but rather reusing them.
	if err := matchUUIDsToID(transaction, uploadedModel); err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, err
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

// UpdateModel encapsulates the GORM functionality for updating a model with its metadata in a transaction
func UpdateModel(uploadedModel *apiTypes.CausalDecisionModel) (int, error) {
	// Begin transaction.
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	// Match all UUIDs in the uploaded model to existing database IDs
	if err := matchUUIDsToID(transaction, uploadedModel); err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, err
	}

	// First, get the existing model with all associations to properly handle removals
	var existingModel apiTypes.CausalDecisionModel
	if err := transaction.
		Preload("Meta").
		Preload("Meta.Updaters").
		Preload("Diagrams").
		Where("meta_id = ?", uploadedModel.Meta.ID).
		First(&existingModel).Error; err != nil {
		transaction.Rollback()
		return http.StatusNotFound, fmt.Errorf("model with UUID %s not found", uploadedModel.Meta.UUID)
	}

	// Clear model diagrams association
	if err := transaction.Model(&existingModel).Association("Diagrams").Clear(); err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not clear model diagrams: %s", err.Error())
	}

	// Clear meta updaters association
	if err := transaction.Model(&existingModel.Meta).Association("Updaters").Clear(); err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not clear meta updaters: %s", err.Error())
	}

	// Before updating the model, we need to make sure the model isn't going to mess with any of
	// the existing associations inside it's components (for example putting to models shouldn'
	// modify parts of a diagram or parts of a user, but rather just modify what diagrams or
	// users are associated with the model itself)
	// UUIDs have already been matched to IDs, so we can just use the IDs to find the existing associations

	// Iterate through all the diagrams with nonzero IDs and reset them to the way they exist in the database
	// to ensure no discrepancies between them as they exist in the database and them as they exist in the model
	for i := range uploadedModel.Diagrams {
		if uploadedModel.Diagrams[i].ID != 0 {
			var existingDiagram apiTypes.Diagram
			// Do nothing on error, and treat it as a new diagram (don't set it to the existing diagram)
			if err := transaction.Where("id = ?", uploadedModel.Diagrams[i].ID).First(&existingDiagram).Error; err == nil {
				uploadedModel.Diagrams[i] = existingDiagram
			}
		}
	}

	// Iterate through all the updaters with nonzero IDs and reset them to the way they exist in the database
	// to ensure no discrepancies between them as they exist in the database and them as they exist in the model
	for i := range uploadedModel.Meta.Updaters {
		if uploadedModel.Meta.Updaters[i].ID != 0 {
			var existingUpdater apiTypes.User
			// Do nothing on error, and treat it as a new updater (don't set it to the existing updater)
			if err := transaction.Where("id = ?", uploadedModel.Meta.Updaters[i].ID).First(&existingUpdater).Error; err == nil {
				uploadedModel.Meta.Updaters[i] = existingUpdater
			}
		}
	}

	// Update the model
	if err := transaction.Save(&uploadedModel).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not update model: %s", err.Error())
	}

	// Commit the transaction
	if err := transaction.Commit().Error; err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not commit transaction: %s", err.Error())
	}

	return http.StatusCreated, nil
}

// CreateCommit encapsulates the GORM functionality for creating a commit in a transaction
func CreateCommit(uploadedCommit *apiTypes.Commit) (int, error) {

	// Begin transaction.
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	// Create the commit in transaction; error out on failure.
	if err := transaction.Create(&uploadedCommit).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not create commit: %s", err.Error())
	}

	// Commit the transaction; error out if commit fails.
	if err := transaction.Commit().Error; err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not commit transaction: %s", err.Error())
	}

	return http.StatusCreated, nil
}
