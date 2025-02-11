package main

import (
	"fmt"
	"opendi/model-hub/api/handlers"
	"os"

	"github.com/gin-gonic/gin"
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
	modelHubAddress := "localhost"
	modelHubPort := "8080"
	// Check to make sure the environment variables are set before using them
	val, ok := os.LookupEnv("OPENDI_MODEL_HUB_ADDRESS")
	if !ok || val == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPENDI_MODEL_HUB_ADDRESS is not set or empty")
		os.Exit(1)
	}
	modelHubAddress = val
	val, ok = os.LookupEnv("OPENDI_MODEL_HUB_PORT")
	if !ok || val == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPENDI_MODEL_HUB_PORT is not set or empty")
		os.Exit(1)
	}
	modelHubPort = val

	router.Run(modelHubAddress + ":" + modelHubPort)
}
