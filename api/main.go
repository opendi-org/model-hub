//
// COPYRIGHT OpenDI
//

package main

import (
	"flag"
	"fmt"
	"opendi/model-hub/api/handlers"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"opendi/model-hub/api/database"

	_ "opendi/model-hub/api/docs"
	"time"
)

func main() {
	// running the server using "go run main.go -e" runs the server in engine mode,
	// which registers only certain routes needed for the CLI
	engMode := flag.Bool("e", false, "Runs the Model Hub in engine mode for the CLI")
	flag.Parse()

	fmt.Println("Starting Model Hub API")
	if *engMode {
		fmt.Println("Running in engine mode")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://129.213.115.50:3000"}, // React frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//import environment variables
	err := godotenv.Load("./config/.env")
	if err != nil {
		fmt.Println("Unable to import environment variables: ", err)
	}

	// Wait for 3 seconds to allow the database to start up before initializing the connection to the database table
	time.Sleep(3 * time.Second)
	ret, err := database.InitializeDBInstance()
	if ret != 0 {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}

	// Insert test data if we are in development mode
	if os.Getenv("DEV_MODE") == "true" {
		database.CreateExampleData()
	}

	//initialize handlers
	modelHandler, _ := handlers.NewModelHandler()
	var authHandler *handlers.AuthHandler
	if !*engMode {
		authHandler, _ = handlers.NewAuthHandler(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
		)
	}

	// TODO fix this logic
	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	//router group for all endpoints related to models
	models := router.Group("/v0/models")
	{
		if *engMode {
			models.GET("", modelHandler.GetModels)
			models.POST("", modelHandler.UploadModel)
			models.GET("/tag/:tag", modelHandler.GetModelByTag)
			models.GET("/commits/:uuid", modelHandler.GetCommitsByModelUUID)
			models.GET("/lineage/:uuid", modelHandler.GetModelLineage)
		} else {
			models.GET("", modelHandler.GetModels)
			models.POST("", modelHandler.UploadModel)
			models.GET("/:uuid", modelHandler.GetModelByUUID)
			models.GET("/tag/:tag", modelHandler.GetModelByTag)
			models.GET("/commits/:uuid", modelHandler.GetCommitsByModelUUID)
			models.GET("/commits/latest/:uuid", modelHandler.GetLatestCommitByModelUUID)
			models.GET("/lineage/:uuid", modelHandler.GetModelLineage)
			models.GET("/children/:uuid", modelHandler.GetModelChildren)
			models.GET("/version/:uuid/:version", modelHandler.GetVersionOfModel)
			models.GET("/search/:type/:name", modelHandler.ModelSearch)
			models.GET("/privacy/:uuid", modelHandler.GetModelPrivacy)
			models.PUT("/privacy/:uuid", modelHandler.PutModelPrivacy)
			models.GET("/transfer/:uuid", modelHandler.GetTransfer)
			models.POST("/transfer/:uuid", modelHandler.PostTransfer)
			models.DELETE("/transfer/:uuid", modelHandler.DeleteTransfer)
		}
	}

	if !*engMode {
		auth := router.Group("/auth")
		{
			auth.GET("/google/login", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.GET("/me", authHandler.GetCurrentUser)
			auth.POST("/logout", authHandler.Logout)
			if os.Getenv("DEV_MODE") == "true" {
				auth.GET("/testlogin", authHandler.TestLogin)
			}
		}
	}

	modelHubAddress := "localhost"
	modelHubPort := "8080"
	if *engMode {
		modelHubPort = "7070" // engine mode should always use 7070
	} else {
		val, ok := os.LookupEnv("OPENDI_MODEL_HUB_ADDRESS")
		if ok && val != "" {
			modelHubAddress = val
		} else {
			// note that value is empty, but just use default
			fmt.Println("Environment variable OPENDI_MODEL_HUB_ADDRESS is not set or empty")
		}
		val, ok = os.LookupEnv("OPENDI_MODEL_HUB_PORT")
		if ok && val != "" {
			modelHubPort = val
		} else {
			// note that value is empty, but just use default
			fmt.Println("Environment variable OPENDI_MODEL_HUB_PORT is not set or empty")
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(modelHubAddress + ":" + modelHubPort)
}
