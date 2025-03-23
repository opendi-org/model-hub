//
// COPYRIGHT OpenDI
//

package handlers

import (
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"

	"github.com/gin-gonic/gin"
	jsondiff "github.com/wI2L/jsondiff"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
}

// AuthHandler struct for handling user login/auth requests
type AuthHandler struct {
}

type LineageHandler struct {
}

// method for getting an instance of ModelHandler
func NewModelHandler() (*ModelHandler, error) {

	return &ModelHandler{}, nil
}

func NewAuthHandler() (*AuthHandler, error) {
	return &AuthHandler{}, nil
}

func NewLineageHandler() (*LineageHandler, error) {
	return &LineageHandler{}, nil
}

// GetModels godoc
// @Summary      Get all models
// @Description  gets all models
// @Tags         models
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/models/ [get]
func (h *ModelHandler) GetModels(c *gin.Context) {
	var models []apiTypes.CausalDecisionModel
	status, models, err := database.GetAllModels()
	if models == nil {
		c.JSON(status, gin.H{"Error": err.Error()})
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, models)
}

// UploadModel godoc
// @Summary      Upload a new model
// @Description  Uploads a causal decision model along with its metadata in a single transaction.
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.CausalDecisionModel  true  "Causal Decision Model Payload"
// @Success      201 {object} apiTypes.CausalDecisionModel "Created model"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      409 {object} gin.H "Conflict: Model with same UUID already exists"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	database.ResetTables()
	database.CreateExampleModels()
	var uploadedModel apiTypes.CausalDecisionModel

	// Bind the JSON payload to the uploaded model struct
	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Call the encapsulated CreateModel method from the database package
	if status, err := database.CreateModelGivenEmail(&uploadedModel); err != nil {
		// Return error based on the CreateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response if model creation is successful
	c.JSON(http.StatusCreated, uploadedModel)
}

// GetModelByUUID godoc
// @Summary      Get model by its uuid
// @Description  gets models using its uuid
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Model not found"
// @Router       /v0/models/{uuid} [get]
func (h *ModelHandler) GetModelByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	status, model, err := database.GetModelByUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the model if found
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, model)
}

// putModel godoc
// @Summary      Update model
// @Description  Updates a causal decision model along with its metadata in a single transaction.
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.CausalDecisionModel  true  "Causal Decision Model Payload"
// @Success      201 {object} apiTypes.CausalDecisionModel "Updated model"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/models/ [put]
func (h *ModelHandler) PutModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	// Bind the JSON payload to the uploaded model struct
	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	status, oldmodel, err := database.GetModelByUUID(uploadedModel.Meta.UUID)

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Update the model before creating the commit so that on a bad
	// put, we don't have to roll back the commit.
	if status, err := database.UpdateModel(&uploadedModel); err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	status, changedModel, err := database.GetModelByUUID(uploadedModel.Meta.UUID)

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	diff, err := jsondiff.Compare(changedModel, oldmodel)

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(500, gin.H{"Error": err.Error()})
		return
	}

	var commit apiTypes.Commit

	commit.CDMUUID = uploadedModel.Meta.UUID
	commit.Diff = diff.String()
	commit.UserUUID = uploadedModel.Meta.Creator.UUID

	if status, err := database.CreateCommit(&commit); err != nil {
		// Return error based on the CreateCommit function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response if model put is successful
	c.JSON(http.StatusCreated, uploadedModel)
}

// GetCommits godoc
// @Summary      Get all commits
// @Description  gets all commits
// @Tags         commits
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/commits/ [get]
func (h *ModelHandler) GetCommits(c *gin.Context) {
	var models []apiTypes.Commit
	status, models, err := database.GetAllCommits()
	if models == nil {
		c.JSON(status, gin.H{"Error": err.Error()})
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, models)
}

// UploadCommit godoc
// @Summary      Upload a new commit
// @Description  Uploads a commit
// @Tags         commits
// @Accept       json
// @Produce      json
// @Param        model  body  apiTypes.Commit  true  "Commit Payload"
// @Success      201 {object} apiTypes.Commit "Created Commit"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      409 {object} gin.H "Conflict: Commit with same UUID already exists"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /v0/commits/ [post]
func (h *ModelHandler) UploadCommit(c *gin.Context) {
	var uploadedCommit apiTypes.Commit

	// Bind the JSON payload to the uploaded commit struct
	if err := c.ShouldBindJSON(&uploadedCommit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Call the encapsulated CreateCommit method from the database package
	if status, err := database.CreateCommit(&uploadedCommit); err != nil {
		// Return error based on the CreateCommit function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return a successful response if commit creation is successful
	c.JSON(http.StatusCreated, uploadedCommit)
}

// UserLogin godoc
// @Summary Login a user
// @Description Logs a user if they exist, or creates a new user
// @Accept json
// @Produce json
// @Param email password
// @Success 200
// @Failure 401 {object} gin.H "Unauthorized"
func (h *AuthHandler) UserLogin(c *gin.Context) {
	//For now, whenever a user logs in, even if the user doesn't exist we just create a new user and log them in.
	email := c.Query("email")
	pass := c.Query("password")
	fmt.Println("EMAIL: " + email)

	status, user, err := database.UserLogin(email, pass)

	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	user.Password = "secret"
	fmt.Println(user)

	// Return the user
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, user)
}

func (h *LineageHandler) GetModelLineage(c *gin.Context) {
	uuid := c.Param("uuid")
	status, lineage, err := database.GetModelLineage(uuid)
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, lineage)
}

func (h *LineageHandler) GetModelChildren(c *gin.Context) {
	uuid := c.Param("uuid")
	status, children, err := database.GetModelChildren(uuid)
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, children)
}

func (h *ModelHandler) ModelSearch(c *gin.Context) {
	fmt.Println("Searching")
	searchType := c.Param("type")
	name := c.Param("name")
	if searchType == "model" {
		fmt.Println("Searching by Model")
		status, models, err := database.SearchModelsByName(name)
		if err != nil {
			c.JSON(status, gin.H{"Error": err.Error()})
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(status, models)
	} else if searchType == "user" {
		fmt.Println("Searching by User")
		status, models, err := database.SearchModelsByUser(name)
		if err != nil {
			c.JSON(status, gin.H{"Error": err.Error()})
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(status, models)
	} else {
		c.JSON(404, gin.H{"Error": "This type of search does not exist"})
		return
	}
}
