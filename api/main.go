package main

import (
	"fmt"
	"opendi/model-hub/api/handlers"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "opendi/model-hub/api/docs"
)

func main() {
	router := gin.Default()

	// Construct the Data Source Name (DSN) for the database connection

	// Check to make sure the environment variables for the database connection are set before using them
	username, ok := os.LookupEnv("OPEN_DI_DB_USERNAME")
	if !ok || username == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_USERNAME is not set or empty")
		os.Exit(1)
	}
	password, ok := os.LookupEnv("OPEN_DI_DB_PASSWORD")
	if !ok || password == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_PASSWORD is not set or empty")
		os.Exit(1)
	}
	hostname, ok := os.LookupEnv("OPEN_DI_DB_HOSTNAME")
	if !ok || hostname == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_HOSTNAME is not set or empty")
		os.Exit(1)
	}
	port, ok := os.LookupEnv("OPEN_DI_DB_PORT")
	if !ok || port == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_PORT is not set or empty")
		os.Exit(1)
	}
	dbname, ok := os.LookupEnv("OPEN_DI_DB_NAME")
	if !ok || dbname == "" {
		// error exit since the value is empty
		fmt.Println("Environment variable OPEN_DI_DB_NAME is not set or empty")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, hostname, port, dbname)

	//initialize handler
	modelHandler, err := handlers.NewModelHandler(dsn)
	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	// Debug, creates a model and meta in the database
	modelHandler.CreateModel()

	//router gruop for all endpoints related to models
	models := router.Group("/models")
	{
		models.GET("/", modelHandler.GetModels) // Use the handler method
	}

	//router group for all endpoints related to models
	model := router.Group("/model")
	{
		model.GET("/:id", modelHandler.GetModelById) // Use the handler method
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

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(modelHubAddress + ":" + modelHubPort)
}
