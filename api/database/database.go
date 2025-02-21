package database


import (
	"fmt"
	"opendi/model-hub/api/apiTypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"github.com/gin-gonic/gin"
	"time"
)


//global db instance
var dbInstance *gorm.DB


//initialize db instance
func InitializeDBInstance(dsn string) (int, error) {
	var err error
	dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return 1, err
	}

	// AutoMigrate all the structs defined in apitypes.go
	err = dbInstance.AutoMigrate(
		&apiTypes.CausalDecisionModel{},
		&apiTypes.Meta{},
		&apiTypes.Diagram{},
		&apiTypes.DiaElement{},
		&apiTypes.CausalDependency{},
	)
	if err != nil {
		return 1, err
	}
	return 0, err

}

//gets singleton db instance 
func GetDBInstance() *gorm.DB {
	return dbInstance
}

//function for getting all models in Go struct  - remember, in Go, public methods have to be capitalized
func GetAllModels(c *gin.Context) (int, []apiTypes.CausalDecisionModel, error) {
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
		Find(&models).Error; err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, models, nil


}


// Example method that creates a sample model in the database
func CreateExampleModel() {
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
		Creator:       "Test Creator",
		CreatedDate:   "2021-07-01",
		Updator:       "Test Updator",
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    1,
		Meta:      meta,
		Diagrams:  nil,
	}

	if err := dbInstance.Create(&model).Error; err != nil {
		fmt.Println("Error creating model: ", err)
	}
}



// CreateModel encapsulates the GORM functionality for creating a model with its metadata in a transaction
func CreateModel(uploadedModel *apiTypes.CausalDecisionModel) (int, error) {
	// Ensure no other model with the same UUID exists.
	var existingMeta apiTypes.Meta
	if err := dbInstance.Where("uuid = ?", uploadedModel.Meta.UUID).First(&existingMeta).Error; err == nil {
		// If there wasn't an error (error is nil), then a meta with the same UUID exists
		return http.StatusConflict, fmt.Errorf("a model with UUID %s already exists", uploadedModel.Meta.UUID)
	}

	// Begin transaction.
	transaction := dbInstance.Begin()
	if transaction.Error != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not begin transaction: %s", transaction.Error.Error())
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
		Where("meta_id = ?", meta.ID).
		First(&model).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("this meta is not associated with a model")
	}

	return http.StatusOK, &model, nil
}
