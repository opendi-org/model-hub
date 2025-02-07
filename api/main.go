package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"opendi/model-hub/api/apiTypes"
)

func main() {
	router := gin.Default()
	router.GET("/models", getModels)
	router.Run("localhost:8080")
}

// Example endpoint that returns an empty CDM, except for it's meta, which is provided with only a creator
func getModels(c *gin.Context) {
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
