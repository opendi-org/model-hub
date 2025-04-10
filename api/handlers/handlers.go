//
// COPYRIGHT OpenDI
//

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	jsonDiffHelpers "opendi/model-hub/api/jsondiffhelpers"
	"strconv" //for applying patches generated with jsondiff

	"github.com/gin-gonic/gin"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
}

// CommitHandler struct for handling commit requests
type CommitHandler struct {
}

// AuthHandler struct for handling user login/auth requests
type AuthHandler struct {
}

// method for getting an instance of ModelHandler
func NewModelHandler() (*ModelHandler, error) {

	return &ModelHandler{}, nil
}

// method for getting an instance of CommitHandler
func NewCommitHandler() (*CommitHandler, error) {

	return &CommitHandler{}, nil
}

func NewAuthHandler() (*AuthHandler, error) {
	return &AuthHandler{}, nil
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
	c.Header("Access-Control-Allow-Origin", "*")
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
	//TODO  - remember to lock database for transacitons that can have race conditions for multiple users!
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

	changedModel, status, err := database.UpdateModelAndCreateCommit(&uploadedModel, oldmodel)
	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	// Return a successful response if model put is
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusCreated, changedModel)
}

// GetCommits godoc
// @Summary      Get all commits
// @Description  gets all commits
// @Tags         commits
// @Produce      json
// @Success      200
// @Failure      500
// @Router       /v0/commits/ [get]
func (h *CommitHandler) GetCommits(c *gin.Context) {
	var models []apiTypes.Commit
	status, models, err := database.GetAllCommits()
	if models == nil {
		c.JSON(status, gin.H{"Error": err.Error()})
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, models)
}

func (h *CommitHandler) GetLatestCommitByModelUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the encapsulated GetModelByUUID function from the database package
	status, commit, err := database.GetLatestCommitForModelUUID(uuid)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the commit if found
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, commit)
}

// doesn't do anything to the database, but just returns the version of the model associated with the commit version
func (h *ModelHandler) GetVersionOfModel(c *gin.Context) {
	strVersion := c.Param("version")
	uuid := c.Param("uuid")
	version, err := strconv.Atoi(strVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	_, latestVersionOfModel, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	status, commit, err := database.GetLatestCommitForModelUUID(uuid)
	if version == 0 && status == http.StatusNotFound {
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(http.StatusOK, latestVersionOfModel)
		return
	}
	if err != nil {
		fmt.Println("Line 221")
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	if version == commit.Version {
		c.JSON(http.StatusOK, latestVersionOfModel)
		return
	}

	if version > commit.Version {
		c.JSON(http.StatusConflict, gin.H{"Error": "Version requested is greater than the latest version"})
	}

	currVersion := commit.Version
	currCommit := commit

	currModelBytes, err := json.Marshal(latestVersionOfModel)

	if err != nil {
		// Return error based on the UpdateModel function response
		fmt.Println("Line 241")
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	//reaches here

	for {
		diff := []byte(currCommit.Diff)
		modified, err := jsonDiffHelpers.ApplyInvertedPatch(currModelBytes, diff)
		if err != nil {
			fmt.Println("Line 252")
			fmt.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//reset variables for next iteration of applying patches
		currModelBytes = modified
		currVersion--
		parentIdStr := currCommit.ParentCommitID

		//if we reach the version we want, return the model
		if currVersion <= version {
			break
		}

		if parentIdStr == "" {
			fmt.Println("Line 268")
			c.JSON(http.StatusInternalServerError, "No parent ID") //if we encounter a null parent id, return error.
			return
		}

		parentId, _ := strconv.ParseInt(parentIdStr, 10, 64)

		_, currCommit, _ = database.GetCommitByID(int(parentId))

	}

	finalModel := apiTypes.CausalDecisionModel{}
	if err := json.Unmarshal(currModelBytes, &finalModel); err != nil {
		fmt.Println("Line 281")
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, finalModel)

}

/* //ERIC - we only needed this for testing.
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
func (h *CommitHandler) UploadCommit(c *gin.Context) {
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
*/

// userLogin godoc
// @Summary      Login a user
// @Description  Either login or create a user
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        user  body  apiTypes.User  true  "User login"
// @Success      201 {object} apiTypes.User "logged in user"
// @Failure      400 {object} gin.H "Bad Request"
// @Failure      500 {object} gin.H "Internal Server Error"
// @Router       /login [put]
func (h *AuthHandler) UserLogin(c *gin.Context) {
	//For now, whenever a user logs in, even if the user doesn't exist we just create a new user and log them in.
	email := c.Query("email")
	pass := c.Query("password")

	status, user, err := database.UserLogin(email, pass)

	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	user.Password = "secret"

	// Return the user
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, user)
}

func (h *ModelHandler) GetModelLineage(c *gin.Context) {
	uuid := c.Param("uuid")
	status, lineage, err := database.GetModelLineage(uuid)
	if err != nil {
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, lineage)
}

func (h *ModelHandler) GetModelChildren(c *gin.Context) {
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
	searchType := c.Param("type")
	name := c.Param("name")
	if searchType == "model" {
		status, models, err := database.SearchModelsByName(name)
		if err != nil {
			c.JSON(status, gin.H{"Error": err.Error()})
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(status, models)
	} else if searchType == "user" {
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

// GetCommitsByModelUUID godoc
// @Summary      Get all commits for a model
// @Description  gets all commits for a specific model by its UUID
// @Tags         commits
// @Produce      json
// @Param        uuid path string true "Model UUID"
// @Success      200
// @Failure      404 {object} gin.H "Commits not found"
// @Router       /v0/commits/model/{uuid} [get]
func (h *CommitHandler) GetCommitsByModelUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	// Call the database function to get all commits for the model
	status, commits, err := database.GetCommitsByModelUUID(uuid)
	if err != nil {
		// If error, return an appropriate response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	// Return the commits if found
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(status, commits)
}
