package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"opendi/model-hub/api/apiTypes"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
	// dependencies such as services, repositories
}

// method for getting an instance of ModelHandler
func NewModelHandler() *ModelHandler {
	return &ModelHandler{}
}

// Example endpoint that returns an empty CDM, except for its meta, which is provided with only a creator
func (h *ModelHandler) GetModels(c *gin.Context) {
	var model = new(apiTypes.CausalDecisionModel)
	model.ID = 1
	model.Schema = "something"
	model.MetaID = 1
	var meta = new(apiTypes.Meta)
	meta.Creator = "Lupito"
	model.Meta = *meta

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, model)
}