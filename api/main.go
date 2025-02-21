//
// COPYRIGHT OpenDI
//


package main

import (
	"fmt"
	"opendi/model-hub/api/handlers"
	"github.com/gin-contrib/cors"
	"os"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"opendi/model-hub/api/database"

	_ "opendi/model-hub/api/docs"
)

func main() {
	router := gin.Default()
	
	router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"}, // React frontend URL
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

	//import environment variables
	err := godotenv.Load()
    if err != nil {
        fmt.Println("Error importing environment variables: ", err)
		os.Exit(1)
    }


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

	ret, err := database.InitializeDBInstance(dsn)
	if ret != 0{
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}

	//initialize handler
	modelHandler, err := handlers.NewModelHandler(dsn)
	// Handle any errors that occur during initialization of the API endpoint handling logic
	if err != nil {
		fmt.Println("Error initializing model handler: ", err)
		os.Exit(1)
	}

	// Debug, creates a model and meta in the database
	database.CreateExampleModel()

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

		models.GET("", modelHandler.GetModels)       // Get all models
		models.GET("/:id", modelHandler.GetModelByUUID) // Get a model by ID
		models.POST("", modelHandler.UploadModel)    // Upload a model
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

	router.Run(modelHubAddress + ":" + modelHubPort)
}
