package main

import (
	"github.com/gin-gonic/gin"
	"opendi/model-hub/api/handlers"
	"fmt"
	"os"
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

	// Get the address and port from environment variables
	// If not set, use default values
	modelHubAddress := "localhost"
	modelHubPort := "8080"
	// Check to make sure the environment variables are set before using them
	val, ok := os.LookupEnv("MODEL_HUB_ADDRESS")
	if !ok || val == "" {
		// error exit since the value is empty
		fmt.Println("MODEL_HUB_ADDRESS is not set or empty")
		os.Exit(1)
	}
	val, ok = os.LookupEnv("MODEL_HUB_PORT")
	if !ok || val == "" {
		// error exit since the value is empty
		fmt.Println("MODEL_HUB_PORT is not set or empty")
		os.Exit(1)
	}
		

	router.Run(modelHubAddress + ":" + modelHubPort)
}
