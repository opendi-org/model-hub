package handlers

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"strconv"
	"time"

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
	meta := apiTypes.Meta{
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
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	c.ShouldBindJSON(&uploadedModel)

	transacation := h.DB.Begin()

	//if what is passed in doesnt aleayd have meta data we need to create meta data

	err := transacation.Create(&uploadedModel.Meta).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	err = transacation.Create(&uploadedModel).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	err = transacation.Commit().Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, uploadedModel)

}

// GetModels godoc
// @Summary      Get model by its id
// @Description  gets models using its id
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        id path int true "Model ID"
// @Success      200
// @Failure      400
// @Failure      404
// @Router       /v0/models/{id} [get]
func (h *ModelHandler) GetModelById(c *gin.Context) {

	var model apiTypes.CausalDecisionModel

	idString := c.Param("id")
	id, err := strconv.Atoi(idString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Updated query with preload for associations
	if err := h.DB.
		Preload("Meta").
		Preload("Diagrams").
		Preload("Diagrams.Meta").
		Preload("Diagrams.Elements").
		Preload("Diagrams.Dependencies").
		Preload("Diagrams.Elements.Meta").
		Preload("Diagrams.Dependencies.Meta").
		First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Model not found"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, model)
}
