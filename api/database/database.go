//
// COPYRIGHT OpenDI
//

package database

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"os"
	"strings"
	"time"

	"github.com/wI2L/jsondiff"
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
		&apiTypes.User{},
		&apiTypes.Transfer{},
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

// initialize db instance with expected tables.
func InitializeDBInstance() (int, error) {

	// Construct the Data Source Name (DSN) for the database connection

	// Check to make sure the environment variables for the database connection are set before using them
	username, ok := os.LookupEnv("OPEN_DI_DB_USERNAME")
	if !ok || username == "" {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_USERNAME is not set or empty")
	}
	password, ok := os.LookupEnv("OPEN_DI_DB_PASSWORD")
	if !ok {
		return 1, fmt.Errorf("environment variable OPEN_DI_DB_PASSWORD is not set")
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
func GetAllModels() ([]apiTypes.CausalDecisionModel, error) {
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
		return nil, err
	}

	return models, nil
}

// GetModelByUUID encapsulates the GORM functionality for getting a model by its UUID
func GetModelByUUID(uuid string) (*apiTypes.CausalDecisionModel, error) {
	var meta apiTypes.Meta

	// Find the meta record with the given UUID.
	if err := dbInstance.Where("uuid = ?", uuid).First(&meta).Error; err != nil {
		return nil, fmt.Errorf("meta with uuid %s not found", uuid)
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
		Preload("Diagrams.Meta.Creator").
		Preload("Diagrams.Meta.Updaters").
		Preload("Diagrams.Elements.Meta.Creator").
		Preload("Diagrams.Elements.Meta.Updaters").
		Preload("Diagrams.Dependencies.Meta.Creator").
		Preload("Diagrams.Dependencies.Meta.Updaters").
		Where("meta_id = ?", meta.ID).
		First(&model).Error; err != nil {
		return nil, fmt.Errorf("this meta is not associated with a model")
	}

	return &model, nil
}

func GetModelByTag(tag string) (*apiTypes.CausalDecisionModel, error) {
	var model apiTypes.CausalDecisionModel
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
		Preload("Diagrams.Meta.Creator").
		Preload("Diagrams.Meta.Updaters").
		Preload("Diagrams.Elements.Meta.Creator").
		Preload("Diagrams.Elements.Meta.Updaters").
		Preload("Diagrams.Dependencies.Meta.Creator").
		Preload("Diagrams.Dependencies.Meta.Updaters").
		Where("JSON_EXTRACT(addons, '$.tag') = ?", tag).First(&model).Error; err != nil {
		return nil, fmt.Errorf("model with tag %s not found", tag)
	}

	return &model, nil
}

func SearchModelsByName(name string) ([]apiTypes.CausalDecisionModel, error) {
	var models []apiTypes.CausalDecisionModel

	if err := dbInstance.
		Joins("JOIN meta ON causal_decision_models.meta_id = meta.id").
		Where("meta.name LIKE ?", "%"+name+"%").
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
		return nil, err
	}

	return models, nil
}

func SearchModelsByUser(username string) ([]apiTypes.CausalDecisionModel, error) {
	var models []apiTypes.CausalDecisionModel

	if err := dbInstance.
		Joins("JOIN users ON users.id = CAST(JSON_EXTRACT(causal_decision_models.addons, '$.ownerID') AS UNSIGNED)").
		Where("users.username LIKE ?", "%"+username+"%").
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
		return nil, err
	}

	return models, nil
}

// / GetModelLineage returns the ancestry of a model given its UUID.
// It retrieves the model and its ancestors in reverse order, starting from the most recent ancestor.
func GetModelLineage(uuid string) ([]apiTypes.CausalDecisionModel, error) {
	modelPtr, err := GetModelByUUID(uuid)
	if err != nil {
		return nil, err
	}

	model := *modelPtr

	var lineage []apiTypes.CausalDecisionModel

	for model.Addons.ParentUUID != "" {
		parentPtr, err := GetModelByUUID(model.Addons.ParentUUID)

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

	return lineage, nil
}

// get the children of this model.
func GetModelChildren(uuid string) ([]apiTypes.CausalDecisionModel, error) {
	var children []apiTypes.CausalDecisionModel
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
		Preload("Diagrams.Meta.Creator").
		Preload("Diagrams.Meta.Updaters").
		Preload("Diagrams.Elements.Meta.Creator").
		Preload("Diagrams.Elements.Meta.Updaters").
		Preload("Diagrams.Dependencies.Meta.Creator").
		Preload("Diagrams.Dependencies.Meta.Updaters").
		Where("JSON_EXTRACT(addons, '$.parentUUID') = ?", uuid).
		Find(&children).Error; err != nil {
		return nil, err
	}

	return children, nil
}

func UpdateModelPrivacyByUUID(uuid string, isPublic bool, shares []apiTypes.Share) error {
	transaction := dbInstance.Begin()

	model, err := GetModelByUUID(uuid)
	if err != nil {
		return err
	}

	model.Addons.IsPublic = isPublic
	model.Addons.Shares = shares
	if err := transaction.Save(&model).Error; err != nil {
		transaction.Rollback()
		return fmt.Errorf("could not update model privacy: %s", err.Error())
	}
	if err := transaction.Commit().Error; err != nil {
		return fmt.Errorf("could not commit transaction: %s", err.Error())
	}
	return nil
}

func GetTransferByModelUUID(uuid string) (*apiTypes.Transfer, error) {
	var transfer apiTypes.Transfer
	if err := dbInstance.Where("cdm_uuid = ?", uuid).First(&transfer).Error; err != nil {
		return nil, fmt.Errorf("transfer on uuid %s not found", uuid)
	}
	return &transfer, nil
}

func GetTransfersByTargetUserID(userID int) ([]map[string]interface{}, error) {
	var transfers []apiTypes.Transfer
	if err := dbInstance.Where("to_user_id = ?", userID).Find(&transfers).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for _, transfer := range transfers {
		model, err := GetModelByUUID(transfer.CDMUUID)
		if err != nil {
			continue
		}

		result := map[string]interface{}{
			"transfer":   transfer,
			"modelName":  model.Meta.Name,
			"modelTag":   model.Addons.Tag,
			"fromUserID": transfer.FromUserID,
		}
		results = append(results, result)
	}

	return results, nil
}

func CreateTransfer(transfer *apiTypes.Transfer) error {
	if err := dbInstance.Create(transfer).Error; err != nil {
		return fmt.Errorf("failed to create transfer: %s", err)
	}
	return nil
}

func DeleteTransfer(transfer *apiTypes.Transfer, accept bool) error {
	// if accepting, change the model owner
	if accept {
		model, err := GetModelByUUID(transfer.CDMUUID)
		if err != nil {
			return err
		}
		transaction := dbInstance.Begin()
		model.Addons.OwnerID = transfer.ToUserID
		if err := transaction.Save(&model).Error; err != nil {
			transaction.Rollback()
			return fmt.Errorf("could not update model owner: %s", err.Error())
		}
		if err := transaction.Commit().Error; err != nil {
			return fmt.Errorf("could not commit transaction: %s", err.Error())
		}
	}
	// delete the transfer request
	if err := dbInstance.Delete(&apiTypes.Transfer{}, transfer.ID).Error; err != nil {
		return fmt.Errorf("failed to delete transfer: %s", err)
	}
	return nil
}

func generateTag(name, version string) string {
	return fmt.Sprintf("%s:%s", strings.ReplaceAll(strings.ToLower(name), " ", "-"), version)
}

// CreateModel encapsulates the GORM functionality for creating a model with its metadata in a transaction
func CreateModel(uploadedModel *apiTypes.CausalDecisionModel, creatorID int) (*apiTypes.CausalDecisionModel, int, error) {
	var count int64
	// keep generating UUIDs until a unique one is found
	for {
		// generate a UUID for the model
		uuid, err := generateUUID()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("could not generate UUID: %s", err.Error())
		}
		uploadedModel.Meta.UUID = uuid

		// ensure no other model with the same UUID exists
		dbInstance.Model(&apiTypes.Meta{}).Where("uuid = ?", uploadedModel.Meta.UUID).Count(&count)
		if count == 0 {
			break
		}
	}

	// make sure the tag is set properly and does not conflict
	uploadedModel.Addons.Tag = generateTag(uploadedModel.Meta.Name, uploadedModel.Meta.Version)
	dbInstance.Model(&apiTypes.CausalDecisionModel{}).
		Where("JSON_EXTRACT(addons, '$.tag') = ?", uploadedModel.Addons.Tag).
		Count(&count)
	if count > 0 {
		return nil, http.StatusConflict, fmt.Errorf("tag %s already exists, version may need to be updated", uploadedModel.Addons.Tag)
	}

	retrievedUser, err := GetUserByEmail(uploadedModel.Meta.Creator.Email)
	if err != nil {
		uploadedModel.Meta.Creator.ID = creatorID
		uploadedModel.Meta.CreatorID = creatorID
	} else {
		uploadedModel.Meta.Creator.ID = retrievedUser.ID
		uploadedModel.Meta.CreatorID = retrievedUser.ID
	}

	// make sure the privacy is set correctly
	uploadedModel.Addons.Shares = nil
	uploadedModel.Addons.IsPublic = false
	uploadedModel.Addons.OwnerID = creatorID

	// set timestamps
	uploadedModel.Meta.CreatedAt = time.Now()
	uploadedModel.Meta.UpdatedAt = time.Now()
	uploadedModel.CreatedAt = time.Now()
	uploadedModel.UpdatedAt = time.Now()

	// begin transactions
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	// match all UUIDs in the model to existing database IDs where possible
	// this will ensure that we are not duplicating pre-existing components
	// but rather reusing them
	if err := matchUUIDsToID(transaction, uploadedModel); err != nil {
		transaction.Rollback()
		return nil, http.StatusInternalServerError, err
	}

	// create meta in transaction, error out on failure
	if err := transaction.Create(&uploadedModel.Meta).Error; err != nil {
		transaction.Rollback()
		return nil, http.StatusInternalServerError, fmt.Errorf("could not create model meta: %s", err.Error())
	}

	uploadedModel.MetaID = uploadedModel.Meta.ID

	// create the model in transaction, error out on failure
	if err := transaction.Create(&uploadedModel).Error; err != nil {
		transaction.Rollback()
		return nil, http.StatusInternalServerError, fmt.Errorf("could not create model: %s", err.Error())
	}

	// commit the transaction, error out if commit fails
	if err := transaction.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("could not commit transaction: %s", err.Error())
	}

	return uploadedModel, http.StatusCreated, nil
}

// UpdateModel encapsulates the GORM functionality for updating a model with its metadata in a transaction. This is a helper method for a PUT to a model.
// currently, for diagrams associated with the model, it creates diagrams that are not already in the database.
// however, for exisitng diagrams, it doesn't change them.
func updateModel(uploadedModel *apiTypes.CausalDecisionModel) (int, error) {
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

	// Update the model meta
	if err := transaction.Save(&uploadedModel.Meta).Error; err != nil {
		transaction.Rollback()
		return http.StatusInternalServerError, fmt.Errorf("could not update model: %s", err.Error())
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

// database method for PUT to a model
// model versions are strings and should be semantic, commit versions are integers and used for internal tracking
func UpdateModelAndCreateCommit(uploadedModel *apiTypes.CausalDecisionModel, oldModel *apiTypes.CausalDecisionModel, updaterID int) (*apiTypes.CausalDecisionModel, int, error) {

	// version should be provided and valid
	if uploadedModel.Meta.Version == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("model version must be provided")
	}
	if uploadedModel.Meta.Version == oldModel.Meta.Version {
		return nil, http.StatusBadRequest, fmt.Errorf("model version must differ from previous version")
	}

	// update the tag
	uploadedModel.Addons = oldModel.Addons
	uploadedModel.Addons.Tag = generateTag(uploadedModel.Meta.Name, uploadedModel.Meta.Version)

	// update the model
	if status, err := updateModel(uploadedModel); err != nil {
		return nil, status, err
	}

	// get the changed model
	changedModel, err := GetModelByUUID(uploadedModel.Meta.UUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	// check diff
	changedModelBytes, err := json.Marshal(changedModel)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	oldModelBytes, err := json.Marshal(oldModel)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	diff, err := jsondiff.CompareJSON(oldModelBytes, changedModelBytes, jsondiff.Invertible())
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if diff.String() == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("no changes made to model")
	}

	jsonData, err := json.Marshal(diff)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// get the latest commit version (internal tracking)
	parent, err := GetLatestCommitForModelUUID(uploadedModel.Meta.UUID)

	// create commit with auto-incremented version
	var commit apiTypes.Commit
	commit.CDMUUID = uploadedModel.Meta.UUID
	commit.Diff = string(jsonData)
	commit.UserID = updaterID
	commit.Version = changedModel.Meta.Version
	commit.CreatedAt = time.Now()
	if err != nil {
		commit.ParentID = -1
	} else {
		commit.ParentID = parent.ID

	}

	if status, err := CreateCommit(&commit); err != nil {
		return nil, status, err
	}

	return changedModel, http.StatusOK, nil
}

// function for getting all commits in Go struct
func GetAllCommits() (int, []apiTypes.Commit, error) {
	var commits []apiTypes.Commit
	// Updated query to preload associated fields
	if err := dbInstance.
		Find(&commits).Error; err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, commits, nil
}

// gets commit by primary key id
func GetCommitByID(id int) (int, *apiTypes.Commit, error) {
	var commit apiTypes.Commit
	err := dbInstance.Where("id = ?", id).First(&commit).Error
	if err != nil {
		return http.StatusNotFound, nil, err
	}
	return http.StatusOK, &commit, nil
}

// GetCommitsByModelUUID returns all commits for a model UUID, ordered by version
func GetCommitsByModelUUID(uuid string) ([]apiTypes.Commit, error) {
	var commits []apiTypes.Commit
	err := dbInstance.Where("cdm_uuid = ?", uuid).Order("parent_id DESC").Find(&commits).Error

	if err != nil || len(commits) == 0 {
		return nil, fmt.Errorf("no commits found for model with UUID %s", uuid)
	}

	return commits, nil
}

// get the latest commit for a model with the given UUID
func GetLatestCommitForModelUUID(uuid string) (*apiTypes.Commit, error) {
	var commit apiTypes.Commit
	err := dbInstance.Where("cdm_uuid = ?", uuid).
		Order("created_at DESC").
		First(&commit).Error

	if err != nil {
		return nil, err
	}
	return &commit, nil
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

// GetUserByID encapsulates the GORM functionality for getting a user by their ID
func GetUserByID(id int) (*apiTypes.User, error) {
	var user apiTypes.User

	// find the user record with the given id
	if err := dbInstance.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user with id %d not found", id)
	}

	return &user, nil
}

func GetUserByEmail(email string) (*apiTypes.User, error) {
	var user apiTypes.User

	// Find the user record with the given ID.
	if err := dbInstance.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	return &user, nil
}

func CreateUser(user apiTypes.User) (*apiTypes.User, error) {
	// Ensure no other user with this email exists
	var count int64
	dbInstance.Model(&apiTypes.User{}).Where("email = ?", user.Email).Count(&count)
	if count > 0 {
		// If a user with the same email exists, return a conflict error.
		return nil, fmt.Errorf("a user with email %s already exists", user.Email)
	}

	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return nil, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
	}

	if err := transaction.Create(&user).Error; err != nil {
		transaction.Rollback()
		return nil, fmt.Errorf("could not create user: %s", err.Error())
	}

	transaction.Commit()
	return &user, nil
}

func FindOrCreateUserFromGoogle(name, email, googleID, picture string) (*apiTypes.User, error) {
	var user apiTypes.User
	if err := dbInstance.Where("google_id = ?", googleID).First(&user).Error; err == gorm.ErrRecordNotFound {
		// didn't find an existing user, make a new one
		newUser := apiTypes.User{
			Username: name,
			Email:    email,
			GoogleID: googleID,
			Picture:  picture,
		}
		return CreateUser(newUser)
	}
	return &user, nil
}

//an example  of what this can do is it can allow upload model to upload a model with a diagram that already existed in the database
//essentially, for any component uploaded to the database it makes sure its associations will be set up correctly, without duplicates .

// matchUUIDsToID recursively iterates through a CDM (or really any CDM component)
// and its nested structures and finds matching UUIDs in the database and updates
// the IDs of the components to match the ID found in the database
// It is designed to work with the structs defined in apitypes.go,
// which as of now are CausalDecisionModel, Meta, Diagram, DiaElement, and CausalDependency.
func matchUUIDsToID(tx *gorm.DB, component any) error {
	// Check if this is a Meta struct and create its users if they don't exist
	if meta, ok := component.(*apiTypes.Meta); ok && meta.UUID != "" {
		var existingMeta apiTypes.Meta
		if err := tx.Where("uuid = ?", meta.UUID).First(&existingMeta).Error; err == nil {
			meta.ID = existingMeta.ID
			if meta.CreatedAt.IsZero() {
				meta.CreatedAt = existingMeta.CreatedAt
			}
		}

		// Match Creator email to ID
		if meta.Creator.Email != "" {
			var existingUser apiTypes.User
			if err := tx.Where("email = ?", meta.Creator.Email).First(&existingUser).Error; err == nil {
				meta.Creator = existingUser
				meta.CreatorID = existingUser.ID
			}
		}

		// Match Updaters emails to IDs
		for i, updater := range meta.Updaters {
			if updater.Email != "" {
				var existingUser apiTypes.User
				if err := tx.Where("email = ?", updater.Email).First(&existingUser).Error; err == nil {
					meta.Updaters[i] = existingUser
				}
			}
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
		if err := tx.Where("meta_id = ?", cdm.Meta.ID).First(&existingModel).Error; err == nil {
			cdm.ID = existingModel.ID
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
		if cdm.Addons.ParentUUID != "" {
			var parentMeta apiTypes.Meta
			if err := tx.Where("uuid = ?", cdm.Addons.ParentUUID).First(&parentMeta).Error; err == nil {
				var parentModel apiTypes.CausalDecisionModel
				if err := tx.Where("meta_id = ?", parentMeta.ID).First(&parentModel).Error; err == nil {
					cdm.Addons.ParentID = &parentModel.ID
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
		if err := tx.Where("meta_id = ?", diagram.Meta.ID).First(&existingDiagram).Error; err == nil {
			diagram.ID = existingDiagram.ID
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
	if element, ok := component.(*apiTypes.DiaElement); ok {
		// First match the meta UUID
		if err := matchUUIDsToID(tx, &element.Meta); err != nil {
			return err
		}
		// Try to find the existing DiaElement in the database
		var existingElement apiTypes.DiaElement
		if err := tx.Where("meta_id = ?", element.Meta.ID).First(&existingElement).Error; err == nil {
			element.ID = existingElement.ID
			if element.CreatedAt.IsZero() {
				element.CreatedAt = existingElement.CreatedAt
			}
		}
		return nil
	}

	// Check if this is a CausalDependency struct and match its Meta UUID to ID
	if dependency, ok := component.(*apiTypes.CausalDependency); ok {
		// First match the meta UUID
		if err := matchUUIDsToID(tx, &dependency.Meta); err != nil {
			return err
		}
		// Try to find the existing CausalDependency in the database
		var existingDependency apiTypes.CausalDependency
		if err := tx.Where("meta_id = ?", dependency.Meta.ID).First(&existingDependency).Error; err == nil {
			dependency.ID = existingDependency.ID
			if dependency.CreatedAt.IsZero() {
				dependency.CreatedAt = existingDependency.CreatedAt
			}
		}
		return nil
	}

	return nil
}

// returns UUID string generated randomly
func generateUUID() (string, error) {
	// Create a byte slice to hold the UUID (16 bytes)
	uuidBytes := make([]byte, 16)

	// Fill the slice with random bytes
	_, err := rand.Read(uuidBytes)
	if err != nil {
		return "", err
	}
	// Format the UUID according to the regex pattern:
	// 8-4-4-4-12 lowercase hexadecimal characters
	uuidStr := fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		uuidBytes[0:4],   // First 4 bytes (8 hex digits)
		uuidBytes[4:6],   // Next 2 bytes (4 hex digits)
		uuidBytes[6:8],   // Next 2 bytes (4 hex digits)
		uuidBytes[8:10],  // Next 2 bytes (4 hex digits)
		uuidBytes[10:16], // Last 6 bytes (12 hex digits)
	)

	return uuidStr, nil
}

// Example method that creates sample models in the database
// creates 2 models, parent and child.
// also creates creators for those models
func CreateExampleData() {
	creator := apiTypes.User{
		ID:       1,
		Username: "creator",
		Email:    "creator@gmail.com",
		GoogleID: "creator-googleid",
	}

	childCreator := apiTypes.User{
		ID:       2,
		Username: "childcreator",
		Email:    "childcreator@gmail.com",
		GoogleID: "childcreator-googleid",
	}

	alternate := apiTypes.User{
		ID:       3,
		Username: "alternate",
		Email:    "alternate@gmail.com",
		GoogleID: "alternate-googleid",
	}

	CreateUser(creator)
	CreateUser(childCreator)
	CreateUser(alternate)

	meta := apiTypes.Meta{
		ID:            1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     creator.ID,
		Creator:       creator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{},
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    1,
		Meta:      meta,
		Addons: apiTypes.Addons{
			OwnerID:  creator.ID,
			Tag:      "test-model:1.0",
			IsPublic: true,
		},
	}

	childMeta := apiTypes.Meta{
		ID:            2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e",
		Name:          "Test Child Model",
		Summary:       "This is a test child model, which I want to search for based only on the summary. The summary is very important.",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     childCreator.ID,
		Creator:       childCreator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{},
		UpdatedDate:   "2021-07-01",
	}

	childModel := apiTypes.CausalDecisionModel{
		ID:        2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Child Schema",
		MetaID:    2,
		Meta:      childMeta,
		Addons: apiTypes.Addons{
			ParentUUID: model.Meta.UUID,
			ParentID:   &model.ID,
			Parent:     &model,
			OwnerID:    childCreator.ID,
			Tag:        "test-child-model:1.0",
			IsPublic:   true,
		},
	}

	privateMeta := apiTypes.Meta{
		ID:            3,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6f",
		Name:          "Private Model",
		Summary:       "This is a private model",
		Documentation: nil,
		Version:       "1.0",
		Draft:         false,
		CreatorID:     creator.ID,
		Creator:       creator,
		CreatedDate:   "2021-07-01",
		Updaters:      []apiTypes.User{},
		UpdatedDate:   "2021-07-01",
	}

	privateModel := apiTypes.CausalDecisionModel{
		ID:        3,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Private Schema",
		MetaID:    3,
		Meta:      privateMeta,
		Addons: apiTypes.Addons{
			OwnerID:  creator.ID,
			Tag:      "private-model:1.0",
			IsPublic: false,
		},
	}

	if err := dbInstance.Create(&model).Error; err != nil {
		fmt.Println("Error creating model: ", err)
	}

	if err := dbInstance.Create(&childModel).Error; err != nil {
		fmt.Println("Error creating child model: ", err)
	}

	if err := dbInstance.Create(&privateModel).Error; err != nil {
		fmt.Println("Error creating private model: ", err)
	}

}
