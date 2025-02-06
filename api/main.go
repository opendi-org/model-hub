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

// Example ednpoint that returns an empty CDM
func getModels(c *gin.Context) {
	var model = new(apiTypes.CausalDecisionModel)
    c.IndentedJSON(http.StatusOK, model)
}