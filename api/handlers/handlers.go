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

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	jsondiff "github.com/wI2L/jsondiff"
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
		//TODO fix this so that if we get an error here, we roll back the update
		// Return error based on the UpdateModel function response
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}

	//TODO  - remember to lock database for transacitons that can have race conditions for multiple users!

	//Let's say that the raw JSON of the original JSON file doesn't contain default values.
	//If we take the raw JSOn, translate it into a Go struct (which definition has default values), change some values (and convert the new struct back to JSON),
	// and then perform a JSON diff between the original JSON and the new JSON,
	//then the diff will think that the default values are part of the JSOn files.

	//this means that if we actually try to apply the diff on the raw JSON, we will get an error because the default values are not in the raw JSON file.

	//so this shows how changing Go structs and then applying the JSON on their raw JSON forms can be a problem. (The standard way is to apply changes to the raw JSON forms instead)
	//However, this is not a problem for our purposes, because we only will aplpy the diff when we convert Go structs to raw JSON - not getting raw JSON from somewhere else.

	changedModelBytes, err := json.Marshal(changedModel)
	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(500, gin.H{"Error": err.Error()})
		return
	}
	oldmodelBytes, err := json.Marshal(oldmodel)
	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(500, gin.H{"Error": err.Error()})
		return
	}

	diff, err := jsondiff.CompareJSON(oldmodelBytes, changedModelBytes, jsondiff.Invertible())

	if err != nil {
		// Return error based on the UpdateModel function response
		c.JSON(500, gin.H{"Error": err.Error()})
		return
	}

	var commit apiTypes.Commit

	commit.CDMUUID = uploadedModel.Meta.UUID

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(diff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	commit.Diff = string(jsonData)
	commit.UserUUID = uploadedModel.Meta.Creator.UUID

	status, parent, err := database.GetLatestCommitForModelUUID(uploadedModel.Meta.UUID)

	//if there's no latest commit, this must be the first.
	if status == http.StatusNotFound {
		commit.ParentCommitID = ""
		commit.Version = 1
	} else if status == http.StatusInternalServerError {
		c.JSON(status, gin.H{"Error": err.Error()})
		return

	} else {
		commit.ParentCommitID = fmt.Sprintf("%d", parent.ID)
		commit.Version = parent.Version + 1
	}

	if status, err := database.CreateCommit(&commit); err != nil {
		// Return error based on the CreateCommit function response
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
	_, latestVersionOfModel, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	_, commit, err := database.GetLatestCommitForModelUUID(uuid)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	//reaches here

	for {
		diff := []byte(currCommit.Diff)
		var patch jsondiff.Patch
		//convert from byte array to patch object
		err = json.Unmarshal(diff, &patch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		invertedPatch, err := jsonDiffHelpers.InvertPatch(patch)
		//apply the inverted patch to the current JSON bytes we have

		//get byte array form of JSON form of inverted ptach
		// Marshal the struct into JSON
		invertedPatchBytes, err := json.Marshal(invertedPatch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		jsonpatchPatch, err := jsonpatch.DecodePatch(invertedPatchBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//apply the patch
		modified, err := jsonpatchPatch.Apply(currModelBytes)

		//reset variables for next iteration of applying patches
		currModelBytes = modified
		currVersion--
		parentIdStr := currCommit.ParentCommitID

		//if we reach the version we want, return the model
		if currVersion <= version {
			break
		}

		if parentIdStr == "" {
			c.JSON(http.StatusInternalServerError, "No parent ID") //if we encounter a null parent id, return error.
			return
		}

		parentId, _ := strconv.ParseInt(parentIdStr, 10, 64)

		_, currCommit, err = database.GetCommitByID(int(parentId))

	}
	fmt.Println(string(currModelBytes))

	finalModel := apiTypes.CausalDecisionModel{}
	if err := json.Unmarshal(currModelBytes, &finalModel); err != nil {
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

// putModel godoc
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
