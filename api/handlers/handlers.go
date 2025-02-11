package handlers

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&apiTypes.CausalDecisionModel{})
	if err != nil {
		return nil, err
	}
	return &ModelHandler{DB: db}, nil
}

// Example endpoint that returns models from the database
func (h *ModelHandler) GetModels(c *gin.Context) {
	var models []apiTypes.CausalDecisionModel
	if err := h.DB.Find(&models).Error; err != nil {
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

	if err := h.DB.Create(&model).Error; err != nil {
		fmt.Println("Error creating model: ", err)
	}
}
