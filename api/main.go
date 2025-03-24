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
		//os.Exit(1)
		//I think the above line should remain commented out, so that
		//the program can still run even if the .env file is not found
		//This is because the .env file is not necessary for the program to run
		//It is only necessary for the program to run in a specific environment
	}

	// Wait for 3 seconds to allow the database to start up before initializing the connection to the database table
	time.Sleep(3 * time.Second)
	//initialize db instance
	ret, err := database.InitializeDBInstance()
	if ret != 0 {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}
	//initialize handler
	modelHandler, err := handlers.NewModelHandler()

	authHandler, err := handlers.NewAuthHandler()

	commitHandler, err := handlers.NewCommitHandler()

	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	// Debug, creates a model and meta in the database
	database.CreateExampleModels()

	//router group for all endpoints related to models
	models := router.Group("/v0/models")
	{
		//Note - CORS headers are cached by default, so if you had a problem with CORS, keep clearing the cache or using new incognito tabs

		//note from Eric - remember to make the ending slashes consistent, or else any non-properly formatted request will redirect causing a CORS violation
		//Gin has built-in automatic redirection for missing slashes.
		//In essence, The initial request (GET /v0/models without a slash) gets redirected. (301 or 307) The browser is told to go to /v0/models/ (with a slash).
		//The browser makes a new request, which must also be checked for CORS
		//CORS is to the user agent (browser), not the server. The server can't tell the browser to ignore CORS.
		//If CORS headers were inherited across redirects, a server could allow an unsafe redirect to a malicious site, exposing private data.
		//Browsers treat redirects as new requests.

		/*
					When the browser follows a redirect (e.g., from /v0/models to /v0/models/), the browser does not automatically send a preflight request for the redirect.
			Instead, the browser treats the redirected request as a new separate request, and it needs to be evaluated for CORS again. This is where the issue arises: if the new request does not include the necessary CORS headers, the browser will block it.



		*/

		/*
			From reddit user toonerer

			It's not the API that's malicious, it's the client.

			Let's say you go to www.your-bamk.com (bamk being a misspelling you as a user typed into your browser), the page could behind the scenes be calling your-bank.com with api-calls, and transfer funds and whatever nasty things you can think of, circumventing any any security measures since it would seem like a real user was interacting with the site.

			Or even worse, a trusted unrelated site could be compromised with a script (from an ad service or similar), and that script could start making calls to your-bank.com without you knowing about it, and if you happened to be logged in from earlier, it would just use those credentials.

			With CORS, your-bank.com would just reject the requests.



		*/

		models.GET("", modelHandler.GetModels)            // Get all models
		models.GET("/:uuid", modelHandler.GetModelByUUID) // Get a model by UUID
		models.POST("", modelHandler.UploadModel)         // Upload a model
		models.PUT("", modelHandler.PutModel)             // Update a model

		models.GET("/lineage/:uuid", modelHandler.GetModelLineage)
		models.GET("/children/:uuid", modelHandler.GetModelChildren)
		models.GET("/modelVersion/:uuid/:version", modelHandler.GetVersionOfModel)
	}

	//router group for all endpoints related to models
	commits := router.Group("/v0/commits")
	{

		commits.GET("", commitHandler.GetCommits) // Get all commits
		commits.GET("/:uuid", commitHandler.GetLatestCommitByModelUUID)
		//commits.POST("", commitHandler.UploadCommit) // Create a commit (for testing)
	}

	//router group for uploading models

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

	router.POST("/login", authHandler.UserLogin)

	router.Run(modelHubAddress + ":" + modelHubPort)
}
