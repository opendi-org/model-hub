//
// COPYRIGHT OpenDI
//

package main

import (
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
	fmt.Println("Starting Model Hub API")

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

	// Insert example data when in development
	if os.Getenv("DEV_MODE") == "true" {
		database.ResetTables()
		database.CreateExampleData()
	}

	//initialize handlers
	modelHandler := handlers.NewModelHandler(false)
	authHandler := handlers.NewAuthHandler(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
	)

	// TODO fix this logic
	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	//router group for all endpoints related to models
	models := router.Group("/v0/models")
	{
		models.GET("", modelHandler.GetModels)
		models.POST("", modelHandler.UploadModel)
		models.GET("/:uuid", modelHandler.GetModelByUUID)
		models.GET("/tag/:tag", modelHandler.GetModelByTag)
		models.GET("/commits/:tag", modelHandler.GetCommitsByModelTag)
		models.GET("/commits/latest/:tag", modelHandler.GetLatestCommitByModelTag)
		models.GET("/lineage/:tag", modelHandler.GetModelLineage)
		models.GET("/children/:tag", modelHandler.GetModelChildren)
		models.GET("/version/:uuid/:version", modelHandler.GetVersionOfModel)
		models.GET("/search/:type/:name", modelHandler.ModelSearch)
		models.GET("/privacy/:tag", modelHandler.GetModelPrivacy)
		models.PUT("/privacy/:tag", modelHandler.PutModelPrivacy)
		models.GET("/transfer/:tag", modelHandler.GetTransfer)
		models.POST("/transfer/:tag", modelHandler.PostTransfer)
		models.DELETE("/transfer/:tag", modelHandler.DeleteTransfer)
	}

	userGroup := router.Group("/v0/user")
	{
		// This results in /v0/user/transfers
		userGroup.GET("/transfers", modelHandler.GetUserPendingTransfers)
	}

	// Diagram-focused endpoints used by frontends to render repo/model/tag graphs.
	diagramHandler := handlers.NewDiagramHandler()

	users := router.Group("/v0/users")
	{
		// GET /v0/users/:userID/repos
		users.GET("/:userID/repos", diagramHandler.GetReposForUser)
	}

	repos := router.Group("/v0/repos")
	{
		// GET /v0/repos/:repoID/models
		repos.GET("/:repoID/models", diagramHandler.GetModelsForRepo)
	}

	diagram := router.Group("/v0/diagram")
	{
		// GET /v0/diagram/repos/:repoID
		diagram.GET("/repos/:repoID", diagramHandler.GetDiagramForRepo)
	}

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

	modelHubAddress := "localhost"
	modelHubPort := "8080"
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

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(modelHubAddress + ":" + modelHubPort)
}
