//
// COPYRIGHT OpenDI
//

package handlers

import (
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	"github.com/gin-gonic/gin"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
}

// method for getting an instance of ModelHandler
func NewModelHandler(dsn string) (*ModelHandler, error) {

	return &ModelHandler{}, nil
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
	status, models, err := database.GetAllModels(c)
	if models == nil {
		c.JSON(status, gin.H{"Error": err.Error()})
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, models)
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

	// Bind the JSON payload to the uploaded model struct
	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Call the encapsulated CreateModel method from the database package
	if status, err := database.CreateModel(&uploadedModel); err != nil {
		// Return error based on the CreateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return 
	}

	// Return a successful response if model creation is successful
	c.JSON(http.StatusCreated, uploadedModel)
}

// GetModelByUUID godoc
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
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	status, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the model if found
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, model)
}
