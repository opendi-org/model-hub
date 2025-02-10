package main

import (
	"github.com/gin-gonic/gin"
	"opendi/model-hub/api/handlers" 
)

func main() {
	router := gin.Default()

	//initialize handler
	modelHandler := handlers.NewModelHandler()

	//router gruop for all endpoints related to models
	models := router.Group("/models")
	{
		models.GET("/", modelHandler.GetModels) // Use the handler method
	}

	router.Run("localhost:8080")
}