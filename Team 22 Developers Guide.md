# OpenDI Model Hub

# Developers Guide

# Development Environment Setup, Project Structure, Extension Guidelines, API Documentation, Build & Testing

# Open DI

# CSC 492 Team 22

# 

# Connor Blumsack

# Matthew Bunch Eric Jun

# Alex Mize Jay Pham  

# North Carolina State University

# Department of Computer Science

# 4/25/2025

# Introduction {#introduction}

---

A guide for sponsors or future development teams who wish to extend, modify, or integrate this project with other systems. It covers information such as environment setup, and application structure.

# Table of Contents {#table-of-contents}

---

[**Introduction	2**](#introduction)

[**Table of Contents	2**](#table-of-contents)

[**System Requirements	3**](#system-requirements)

[Required Tools	3](#required-tools)

[**Familiarization with System	3**](#familiarization-with-system)

[**Development Environment	3**](#development-environment)

[Installation Instructions	3](#installation-instructions)

[Setting Up the Project	3](#setting-up-the-project)

[Running the Project	5](#running-the-project)

[Running Unit Tests	5](#running-unit-tests)

[**Project Structure	6**](#project-structure)

[Notable Directories & Files	7](#notable-directories-&-files)

# 

# System Requirements {#system-requirements}

---

Software, libraries, dependencies, and tools required to develop and run the web application.

## Required Tools {#required-tools}

Programming Language(s): Go  
Framework(s): React, Swaggo  
Database System(s): MySQL  
Other Dependencies: Node.js, Docker, NPM  
OS Compatibility: Windows, macOS, Linux

# Familiarization with System {#familiarization-with-system}

---

We suggest understanding the concept of **Provenance** and how it is implemented first. 

# Development Environment {#development-environment}

---

## Installation Instructions {#installation-instructions}

Install `Node.js` with `npm`: [https://nodejs.org/en/download](https://nodejs.org/en/download)  
Install `Go`: [https://go.dev/doc/install](https://go.dev/doc/install)  
Install `Docker`: [https://docs.docker.com/desktop/setup/install/windows-install/](https://docs.docker.com/desktop/setup/install/windows-install/)  
Install `MySQL Workbench`: [https://dev.mysql.com/downloads/workbench/](https://dev.mysql.com/downloads/workbench/)

## Setting Up the Project {#setting-up-the-project}

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

```
# .env
OPEN_DI_DB_USERNAME=root
OPEN_DI_DB_PASSWORD=password
OPEN_DI_DB_HOSTNAME=localhost
OPEN_DI_DB_PORT=3306
OPEN_DI_DB_NAME=openDI_modelhub_dev
OPENDI_MODEL_HUB_ADDRESS=localhost
OPENDI_MODEL_HUB_PORT=8080
```

8. Create database by running `createDB.sql` located in the *api* directory

## Running the Project {#running-the-project}

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
```

## 

## Running Unit Tests {#running-unit-tests}

1. Navigate to the directory containing the .go file to test

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

# Project Structure {#project-structure}

---

**Note:** Some files are omitted from the file structure below

```
/model-hub
├── api/
│   ├── apiTypes/
│   ├── config/
│   ├── database/
│   ├── docs/
│   ├── handlers/
│   ├── jsondiffhelpers/
│   ├── test_files/
│   ├── .env
│   ├── createDB
│   ├── Dockerfile
│   ├── go
│   └── main
├── frontend/
│   └── model-hub/
│       ├── src/
│       │   ├── components/
│       │   ├── pages/
│       │   ├── App.js
│       │   ├── config.js
│       │   ├── logo.png
│       │   ├── opendi-icon.png
│       ├── Dockerfile
│       └── prod
├── compose.prod
└── compose	
```

## Notable Directories & Files {#notable-directories-&-files}

`apiTypes/`: Contains API structure for GORM.  
`database/`: Contains database interaction methods to interact with GORM.  
`docs/`: Contains Swaggo documentation.  
`handlers/`: Contains API functions.  
`components/`: Contains reusable frontend components.  
`jsondiffhelpers/`: Contains functionality for retrieving and performing JSON diffs.   
`pages/`: Contains frontend web pages.

# Extend the Project

---

For instructions on how to extend the project, take a look at:  
Suggestions for Future Teams in our FPR.