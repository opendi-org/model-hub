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
		}

		// TODO: Change this so we no longer create a new user if the UUID is not found
		// Right now this is just a workaround to create a new user, but in the future when
		// we have a way to properly create users, we should not do this, and instead if there
		// is no user with the UUID, we should error out and not create a new user.

		// Resolve Creator UUID
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

		// Resolve Updaters UUIDs
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
		}
		return nil
	}

	// Check if this is a Commit struct and match its (supposedly unique) pair [ParentCommitID, CDMUUID] to ID
	if commit, ok := component.(*apiTypes.Commit); ok {
		// Try to find the existing Commit in the database
		var existingCommit apiTypes.Commit
		if err := tx.Where("parent_commit_id = ? AND cdm_uuid = ?", commit.ParentCommitID, commit.CDMUUID).First(&existingCommit).Error; err == nil {
			commit.ID = existingCommit.ID
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

	// Resolve all UUIDs in the model to existing database IDs where possible
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
	// Before updating the Meta record, first fetch the existing Meta to get its ID
	var existingMeta apiTypes.Meta
	if err := transaction.
		Preload("Updaters").
		Preload("Creator").
		Where("uuid = ?", uploadedModel.Meta.UUID).
		First(&existingMeta).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not find existing meta: %s", err.Error())
	}

	// Set the ID so GORM knows this is an update, not an insert
	uploadedModel.Meta.ID = existingMeta.ID
	// Check to see if created at is zero, if so, set it to the time in the existing record
	if uploadedModel.Meta.CreatedAt.IsZero() {
		uploadedModel.Meta.CreatedAt = existingMeta.CreatedAt
	}
	// Merge in the updaters from the existing record
	// Iterate through the existing updaters and add them to the new updaters if they are not already there
	for _, existingUpdater := range existingMeta.Updaters {
		updaterExists := false
		for _, newUpdater := range uploadedModel.Meta.Updaters {
			if existingUpdater.UUID == newUpdater.UUID {
				updaterExists = true
				break
			}
		}
		if !updaterExists {
			uploadedModel.Meta.Updaters = append(uploadedModel.Meta.Updaters, existingUpdater)
		}
	}

	// Now the save will update the existing record instead of trying to insert a new one
	if err := transaction.Save(&uploadedModel.Meta).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not update model meta: %s", err.Error())
	}

	// updates the model
	if err := transaction.Save(&uploadedModel).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not update model: %s", err.Error())
	}

	// Commit the transaction; error out if commit fails.
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
