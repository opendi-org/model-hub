OpenDI Model Hub - Developer's Guide

Text in Purple is taken from the previous team’s developer’s guide

# System Requirements

---

Software, libraries, dependencies, and tools required to develop and run the web application.

## Required Tools

Programming Language(s): Go  
Framework(s): React, Swaggo  
Database System(s): MySQL  
Other Dependencies: Node.js, Docker, NPM  
OS Compatibility: Windows, macOS, Linux

## Installation Instructions

Install `Node.js` with `npm`: [https://nodejs.org/en/download](https://nodejs.org/en/download)  
Install `Go`: [https://go.dev/doc/install](https://go.dev/doc/install)  
Install `Docker`: [https://docs.docker.com/desktop/setup/install/windows-install/](https://docs.docker.com/desktop/setup/install/windows-install/)
Install `MySQL Workbench`: [https://dev.mysql.com/downloads/workbench/](https://dev.mysql.com/downloads/workbench/)

## Setting Up the Project

1. Clone the repository:

```
$ git clone [repository-url]
```

2. Navigate to the frontend directory:

```
$ cd model-hub/frontend/model-hub
```

3. Install dependencies:

```
$ npm install
```

4. Navigate to the api directory:

```
$ cd ..
$ cd ..
$ cd api/
```

5. Download library dependencies. 

```
go mod tidy
```

6. Install Swaggo:

```
$ go install github.com/swaggo/swag/cmd/swag@latest
$ swag init --parseDependency --parseInternal --parseDepth 1
```

7. Create a copy of .env-example and rename it to .env in the config directory:


# .env
OPEN_DI_DB_USERNAME=root
OPEN_DI_DB_PASSWORD=password
OPEN_DI_DB_HOSTNAME=localhost
OPEN_DI_DB_PORT=3306
OPEN_DI_DB_NAME=openDI_modelhub_dev
OPENDI_MODEL_HUB_ADDRESS=localhost
OPENDI_MODEL_HUB_PORT=8080
DEV_MODE=false
JWT_SECRET=secret
GOOGLE_CLIENT_ID=clientid
GOOGLE_CLIENT_SECRET=clientsecret
```

8. Create database by running `createDB.sql` located in the *api* directory

## Running the Project

**(Recommended)**

1. Ensure Docker is running   
2. In the root directory

```
$ docker compose up
```

**OR**

1. In the *api* directory:

```
$ go run main.go
```

2. In the *frontend/model-hub* directory: inject this environment variable – `REACT_APP_API_URL=http://localhost:8080 –` into npm runtime and run `npm start`

    An example for Git Bash is:

```
$ REACT_APP_API_URL=http://localhost:8080 npm start

## Running Unit Tests

1. Navigate to the `api` directory

**Running All Tests**

```
# Run all tests in all subdirectories
$ cd api
$ go test ./...

# Run tests for a specific package
$ go test ./handlers_test    # Run all handler tests
$ go test ./database          # Run database tests
```

**Running Specific Test Functions**

```
# Run a specific test function
$ go test ./handlers_test -run TestGetModels

# Run all tests matching a pattern (e.g., all auth tests)
$ go test ./handlers_test -run TestAuth

# Run all tests matching a pattern (e.g., all model tests)
$ go test ./handlers_test -run TestModel

# Note: Individual test files cannot be tested directly in Go,
# but you can test specific functions using the -run flag
```

**Without Coverage Reports**

```
$ go test
$ //for testing all subdirectories in the current directory, 
$ go test ./...
```

**With Coverage Reports**

```
$ go test -v -coverprofile cover.out
$ go tool cover -html=cover.out
```

# Project Structure

---

```
.
└── /model-hub/
    ├── api/
    │   ├── apiTypes/
    │   │   └── apiTypes.go
    │   ├── config/
    │   │   ├── .env-example
    │   │   └── .env.test-example
    │   ├── database/
    │   │   ├── database.go
    │   │   └── database_test.go
    │   ├── docs/
    │   │   ├── docs.go
    │   │   ├── swagger.json
    │   │   └── swagger.yaml
    │   ├── handlers/
    │   │   ├── auth.go
    │   │   ├── helpers.go
    │   │   ├── model.go
    │   │   └── transfer.go
    │   ├── handlers_test/
    │   │   ├── auth_test.go
    │   │   ├── model_test.go
    │   │   ├── test_helpers.go
    │   │   └── transfer_test.go
    │   ├── jsondiffhelpers/
    │   │   └── jsondiffhelpers.go 
    │   ├── test_files/
    │   │   └── //Various files for testing
    │   ├── testutils/
    │   │   └── testutils.go 
    │   ├── createDB.sql //Creates the databases
    │   ├── Dockerfile
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   └── prod.Dockerfile
    ├── frontend/
    │   └── model-hub/
    │       ├── public/
    │       │   ├── favicon.ico
    │       │   ├── index.html
    │       │   ├── logo192.png
    │       │   ├── logo512.png
    │       │   ├── manifest.json
    │       │   └── robots.txt
    │       ├── src/
    │       │   ├── components/
    │       │   │   ├── JsonDiffViewer.js
    │       │   │   ├── JsonPatchViewer.js
    │       │   │   ├── ModelMinicard.js
    │       │   │   ├── Navbar.js
    │       │   │   └── NavbarElements.js
    │       │   ├── context/
    │       │   │   └── UserContext.js
    │       │   ├── pages/
    │       │   │   ├── AuthCallback.js
    │       │   │   ├── downloadPage.js
    │       │   │   ├── index.js
    │       │   │   ├── login.js
    │       │   │   ├── modelPage.js
    │       │   │   ├── search.js
    │       │   │   ├── uploadPage.js
    │       │   │   └── user.js
    │       │   ├── App.js
    │       │   ├── App.test.js
    │       │   ├── config.js
    │       │   ├── index.css
    │       │   ├── logo.svg
    │       │   ├── opendi-icon.png
    │       │   ├── reportWebVitals.js
    │       │   ├── setupTests.js
    │       │   └── Theme.js
    │       ├── .gitignore
    │       ├── DockerFile
    │       ├── prod.Dockerfile
    │       └── README
    ├── media/
    │   ├── developer
    │   └── user/
    │       └── image1.pmg-image23.png
    ├── setup/
    │   └── windows/
    │       ├── instructions.txt
    │       └── set-env-example.ps1
    ├── .gitignore
    ├── cli.py
    ├── compose.prod.yaml
    ├── compose.yaml
    ├── coverage
    ├── deployment.md
    ├── Developers Guide.md
    ├── User Manual.md
    ├── LICENSE
    └── README
```

## Notable Directories And Files
api\main.go: main.go is the file that manages the various API endpoint fetches that are created, as well as starting up the API when you want to run the server. The fetches are stored in various router groups, which take the various requests and routes them to the correct handler function. Currently, the system contains three route handlers: userGroup, models, and auth. However, more can be added if needed in the future. In order to add a new router group, define it with router.Group, with the argument being the base path that the urls inside of it begin with.

api\apiTypes: The apiTypes directory contains apiTypes.go, which defines the various struct types present inside the model hub. Each struct definition contains various fields that can then be referenced when the struct is created in one of the other backend files such as handlers.go. These fields are defined in the format ‘name type json’, and if needed it is simple to modify the fields inside of the structs or create new structs.  

api\handlers: The handlers directory contains handlers.go and handlers_test.go, which contain the code for handling API endpoint fetches. Any effect that uses an API endpoint has an associated function within handlers.go, such as getting users in the frontend, searching for a model, or uploading new files. An important thing to remember is that handlers.go does not directly interact with the database, instead working with structs from apiTypes and then either filling the structs with data from database.go or sending filled structs to database.go to be stored in the SQL database. The other file in the directory, handlers_test.go, contains the tests for the different functions in handlers.go. In order to create a new test, create a function with the argument ‘t *testing.T’. Calls to database.go can be made in order to reset the system to an initial state. Afterwards, structs from apiTypes can be created with test data inside of them, and then inserted into the database. Once the data is created, http.NewRequest can be used to simulate the handler receiving a request from main.go, and the response can be checked with assert.

api\database: The database directory contains database.go and database_test,go, which handle the link between the database and the rest of the system. database.go contains an instance of the database and is in charge of initializing the database on startup, as well as containing functions for the various DB Queries that are used by other files like handlers. database_test.go contains the tests for database. In order to create a test, create a function with the argument ‘t *testing.T’. the ResetTables() function can be used to get a blank slate for testing, and functions from database can be tested by creating structs from apiTypes and filling them with json data to use as arguments. If the test conditions are not met, use the t.Fatalf function to fail the test.

frontend\model-hub\src\pages: The pages directory contains the different files that are displayed on the frontend. Frontend pages must be coded in react, and contain a mix of html and js. If new fronted locations need to be created, they will be stored here. In order to redirect to a newly created frontend page, it must be linked in frontend\model-hub\app\App.js inside of the Routes html block.

## Notable API Endpoints

There are many API Endpoints that are present in the system, and they can be viewed in the file api\main.go. Some of these Endpoints are more important than others, with some notable ones to know being listed here.
GET /auth/me: Reads the auth token present in the system and returns its associated user from the database. This is used across the frontend to load the current user into the pages.
GET /auth/google/login: Starts the Google OAuth flow, redirects the user to the GoogleOAuth consent screen. This is used in the login section of the navbar.
GET /v0/models/:uuid: takes a UUID in the body and returns the model in the database that has the same UUID. This is used in the search page for models.
POST /vo/models/: This is used to create a model. A JSON model is stored in the body of the request, and if it is valid a new model is initialized and stored in the database. This is used in the upload page of the model hub to upload models.
GET /v0/user/transfers: Gets a list of models waiting to be transferred to the current user. This is used in the profile page to check if a user has pending transfers.
POST /v0/models/transfer/:tag: Takes a model tag and 2 arguments, the owner ID of the model and the potential new owner ID. Creates a model transfer request in the transfers table of the database, which contains the model that is being transferred and the 2 ids. This is used in the ownership section of the model page.

# The Command Line Interface (CLI)

The Command Line Interface can be used to interact with the system without needing to act through the frontend. Currently, the command line supports the following actions:
Log in and log out
Initiate a Model
Pull a model from the remote repository
Commit changes to a model
Push a model to the remote repository
Get a model’s lineage
Get a model’s commits

## CLI Installation
The Command Line Interface executable download can be accessed from the model hub frontend. In order to download the CLI:
Access the model hub frontend
Log in to an account
Click the ‘Download’ button on the navbar
Press the ‘Download CLI’ button on the new page
Place the executable file into your working directory

## Using the CLI

In order to set up the CLI, follow these instructions while in your working directory:
Use the command ‘opendi-cli.exe set-url http://opendi-modelhub.org’ to set the CLI remote url.
Inside the download page for the CLI, there is a section containing the authentication token for your account. Use the command ‘opendi-cli.exe set-token YOUR_TOKEN_HERE’ to set your authentication token in the system.

Once the CLI is set up, you can now use commands by inputting ‘opendi-cli.exe’ and then a command and its arguments. The download page contains a list of commands, and you can append use the -h command to a command to see how each command functions.

