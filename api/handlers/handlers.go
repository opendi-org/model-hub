//
// COPYRIGHT OpenDI
//

package handlers

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"time"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
	DB *gorm.DB
}

// method for getting an instance of ModelHandler
func NewModelHandler(dsn string) (*ModelHandler, error) {
	var tries = 0
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	for err != nil && tries < 5 {
		time.Sleep(5 * time.Second)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		tries++
	}

	// AutoMigrate all the structs defined in apitypes.go
	err = db.AutoMigrate(
		&apiTypes.CausalDecisionModel{},
		&apiTypes.Meta{},
		&apiTypes.Diagram{},
		&apiTypes.DiaElement{},
		&apiTypes.CausalDependency{},
	)
	if err != nil {
		return nil, err
	}

	return &ModelHandler{DB: db}, nil
}

// GetModels godoc
// @Summary      Get all models
// @Description  gets all models
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/models/ [get]
func (h *ModelHandler) GetModels(c *gin.Context) {
	var models []apiTypes.CausalDecisionModel
	// Updated query to preload associated fields
	if err := h.DB.
		Preload("Meta").
		Preload("Diagrams").
		Preload("Diagrams.Meta").
		Preload("Diagrams.Elements").
		Preload("Diagrams.Dependencies").
		Preload("Diagrams.Elements.Meta").
		Preload("Diagrams.Dependencies.Meta").
		Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, models)
}

// Example endpoint that creates a model in the database
// This endpoint doesn't actually use the request body to create the model,
// it just creates a model with a hard-coded Schema and Meta
func (h *ModelHandler) CreateModel() {
	docJSON := `{
		"content": "This CDD was authored by Dr. Lorien Pratt.\nSource: https://www.lorienpratt.com/a-framework-for-how-data-informs-decisions/\n\nAdapted for OpenDI schema compliance by Isaac Kellogg.",
		"MIMEType": "text/plain"
	}`

	docRaw := json.RawMessage(docJSON)

	meta := apiTypes.Meta{
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UUID:          "1234-5678-9101",
		Name:          "Test Model",
		Summary:       "This is a test model",
		Documentation: docRaw,
		Version:       "1.0",
		Draft:         false,
		Creator:       "Test Creator",
		CreatedDate:   "2021-07-01",
		Updator:       "Test Updator",
		UpdatedDate:   "2021-07-01",
	}

	model := apiTypes.CausalDecisionModel{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Schema:    "Test Schema",
		MetaID:    1,
		Meta:      meta,
		Diagrams:  nil,
	}

	if err := h.DB.Create(&model).Error; err != nil {
		fmt.Println("Error creating model: ", err)
	}
}

// UploadModel godoc
// @Summary      Upload a new model
// @Description  Uploads a causal decision model along with its metadata in a single transaction.
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.CausalDecisionModel  true  "Causal Decision Model Payload"
// @Success      201 {object} apiTypes.CausalDecisionModel "Created model"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      409 {object} gin.H "Conflict: Model with same UUID already exists"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Get the uploaded model's meta info.
	meta := uploadedModel.Meta

	// Ensure no other model with the same UUID exists.
	var existingMeta apiTypes.Meta
	if err := h.DB.Where("uuid = ?", meta.UUID).First(&existingMeta).Error; err == nil {
		// If there wasn't an error (error is nil), then a meta with the same UUID exists
		c.JSON(http.StatusConflict, gin.H{"Error": "A model with UUID " + meta.UUID + " already exists"})
		return
	}

	// Begin transaction.
	transaction := h.DB.Begin()
	if transaction.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": transaction.Error.Error()})
		return
	}

	// Create meta in transaction; error out on failure.
	if err := transaction.Create(&uploadedModel.Meta).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Create the model in transaction; error out on failure.
	if err := transaction.Create(&uploadedModel).Error; err != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Commit the transaction; error out if commit fails.
	if err := transaction.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, uploadedModel)
}

// GetModels godoc
// @Summary      Get model by its uuid
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Router       /v0/models/{uuid} [get]
func (h *ModelHandler) GetModelByUUID(c *gin.Context) {

	var meta apiTypes.Meta
	uuid := c.Param("uuid")
	// Find the meta record with the given uuid
	if err := h.DB.Where("uuid = ?", uuid).First(&meta).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Meta with uuid " + uuid + " not found"})
		return
	}

	var model apiTypes.CausalDecisionModel
	// Find the model that has the found meta record, preloading associated fields
	// This should only error out at this point if the meta is associated with something other than a model,
	// like a diagram or a diagram element
	if err := h.DB.
		Preload("Meta").
		Preload("Diagrams").
		Preload("Diagrams.Meta").
		Preload("Diagrams.Elements").
		Preload("Diagrams.Dependencies").
		Preload("Diagrams.Elements.Meta").
		Preload("Diagrams.Dependencies.Meta").
		Where("meta_id = ?", meta.ID).
		First(&model).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "This meta is not associated with a model"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, model)
}
