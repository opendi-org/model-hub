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

	"github.com/gin-gonic/gin"
)

// ModelHandler struct for handling model requests
type ModelHandler struct {
	engMode bool
}

// NewModelHandler method for getting an instance of ModelHandler
func NewModelHandler(engMode bool) *ModelHandler {
	return &ModelHandler{
		engMode: engMode,
	}
}

// GetModels godoc
// @Summary      Get all models
// @Description  Gets all models
// @Tags         models
// @Produce      json
// @Success      200  {array}   gin.H  "List of models"
// @Failure      500  {object}  gin.H  "Internal server error"
// @Router       /v0/models/ [get]
func (h *ModelHandler) GetModels(c *gin.Context) {
	models, err := database.GetAllModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	filteredModels := filterModelsForActingUser(actingUserID, models)

	result := make([]gin.H, len(filteredModels))
	for i, model := range filteredModels {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// GetModelByUUID godoc
// @Summary      Get model by its UUID
// @Description  Get model by its UUID
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid  path  string  true  "Model UUID"
// @Success      200   {object}  gin.H  "Model"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/{uuid} [get]
func (h *ModelHandler) GetModelByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	model, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	hasAccess := checkIfUserHasAccessToModel(actingUserID, *model)
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	c.JSON(http.StatusOK, addAddonsFields(*model))
}

// GetModelByTag godoc
// @Summary      Get model by its tag
// @Description  Get model by its tag
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag   path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Model"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/tag/{tag} [get]
func (h *ModelHandler) GetModelByTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	hasAccess := checkIfUserHasAccessToModel(actingUserID, *model)
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	c.JSON(http.StatusOK, addAddonsFields(*model))
}

// UploadModel godoc
// @Summary      Creates or updates a model
// @Description  Creates or updates a model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body  object  true  "Model"
// @Success      200    {object}  gin.H  "Model created or updated"
// @Failure      400    {object}  gin.H  "Bad request"
// @Failure      401    {object}  gin.H  "Unauthorized"
// @Failure      403    {object}  gin.H  "Forbidden"
// @Failure      500    {object}  gin.H  "Internal server error"
// @Router       /v0/models [post]
func (h *ModelHandler) UploadModel(c *gin.Context) {
	var uploadedModel apiTypes.CausalDecisionModel

	if err := c.ShouldBindJSON(&uploadedModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var resultModel *apiTypes.CausalDecisionModel
	var status int

	if uploadedModel.Meta.UUID != "" {
		// either the model already exists or the user has provided one (which we should ignore and create another one)
		var retrievedModel *apiTypes.CausalDecisionModel
		if retrievedModel, err = database.GetModelByUUID(uploadedModel.Meta.UUID); err != nil {
			// the user provided a UUID that does not match any model in the system, make a new one and ignore the old UUID
			resultModel, status, err = database.CreateModel(&uploadedModel, actingUserID)
		} else {
			// we found a model with a matching UUID, make sure the user has permissions to update
			canWrite := retrievedModel.Addons.OwnerID == actingUserID
			actingUser, _ := database.GetUserByID(actingUserID)
			if !canWrite {
				for _, share := range retrievedModel.Addons.Shares {
					if share.Email == actingUser.Email && share.Level == "write" {
						canWrite = true
						break
					}
				}
			}
			if !canWrite {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
				return
			}
			// user has permissions to udpate this model, so create a commit
			resultModel, status, err = database.UpdateModelAndCreateCommit(&uploadedModel, retrievedModel, actingUserID)
		}
	} else {
		// a UUID was not provided, we assume this means we are creating a new model
		resultModel, status, err = database.CreateModel(&uploadedModel, actingUserID)
	}
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addAddonsFields(*resultModel))
}

// TODO adjust this so it looks at the version and the first part of the tag separately
// GetVersionOfModel godoc
// @Summary      Get version of model
// @Description  Get version of model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        uuid     path  string  true  "Model UUID"
// @Param        version  path  string  true  "Version number"
// @Success      200      {object}  gin.H  "Model version details"
// @Failure      404      {object}  gin.H  "Version not found"
// @Failure      500      {object}  gin.H  "Internal server error"
// @Router       /v0/models/version/{uuid}/{version} [get]
func (h *ModelHandler) GetVersionOfModel(c *gin.Context) {
	version := c.Param("version")
	uuid := c.Param("uuid")

	// get latest version of model
	latestVersionOfModel, err := database.GetModelByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	hasAccess := checkIfUserHasAccessToModel(actingUserID, *latestVersionOfModel)
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	if latestVersionOfModel.Meta.Version == version {
		c.JSON(http.StatusOK, addAddonsFields(*latestVersionOfModel))
		return
	}

	// get the latest commit; if there isnt one, we are working with a model that has not been updated
	latestCommit, err := database.GetLatestCommitForModelUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	currCommit := latestCommit
	currModelBytes, _ := json.Marshal(latestVersionOfModel)

	// we need to apply the diff to the model in reverse order, so we start with the latest commit and go backwards
	for {
		currModelBytes, err = jsonDiffHelpers.ApplyInvertedPatch(currModelBytes, []byte(currCommit.Diff))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// if we reach the version we want, return the model
		if currCommit.Version == version {
			break
		}
		if currCommit.ParentID == -1 {
			break
		}

		_, currCommit, _ = database.GetCommitByID(currCommit.ParentID)
	}

	finalModel := apiTypes.CausalDecisionModel{}
	if err := json.Unmarshal(currModelBytes, &finalModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if finalModel.Meta.Version != version {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model does not have version: %s", version)})
		return
	}
	c.JSON(http.StatusOK, addAddonsFields(finalModel))

}

// GetModelLineage godoc
// @Summary      Get model lineage
// @Description  Get model lineage
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Model lineage"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/lineage/{tag} [get]
func (h *ModelHandler) GetModelLineage(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	lineage, err := database.GetModelLineage(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	filteredLineage := filterModelsForActingUser(actingUserID, lineage)

	result := make([]gin.H, len(filteredLineage))
	for i, model := range filteredLineage {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// GetModelChildren godoc
// @Summary      Get model children
// @Description  Gets models using its UUID
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {array}   gin.H  "List of child models"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/children/{tag} [get]
func (h *ModelHandler) GetModelChildren(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	children, err := database.GetModelChildren(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	filteredChildren := filterModelsForActingUser(actingUserID, children)

	result := make([]gin.H, len(filteredChildren))
	for i, model := range filteredChildren {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// ModelSearch godoc
// @Summary      Search for models
// @Description  Search for models by name or user
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        type  path  string  true  "Search type (model or user)"
// @Param        name  path  string  true  "Search name"
// @Success      200   {array}   apiTypes.CausalDecisionModel  "List of models"
// @Failure      404   {object}  gin.H  "Models not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/search/{type}/{name} [get]
func (h *ModelHandler) ModelSearch(c *gin.Context) {
	searchType := c.Param("type")
	name := c.Param("name")

	var models []apiTypes.CausalDecisionModel
	var err error
	switch searchType {
	case "model":
		models, err = database.SearchModelsByName(name)
	case "user":
		models, err = database.SearchModelsByUser(name)
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "This type of search does not exist"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	filteredModels := filterModelsForActingUser(actingUserID, models)

	result := make([]gin.H, len(filteredModels))
	for i, model := range filteredModels {
		result[i] = addAddonsFields(model)
	}

	c.JSON(http.StatusOK, result)
}

// GetCommitsByModelTag godoc
// @Summary      Get all commits for a model
// @Description  Get all commits for a model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {array}   gin.H  "List of commits"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/commits/{tag} [get]
func (h *ModelHandler) GetCommitsByModelTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	hasAccess := checkIfUserHasAccessToModel(actingUserID, *model)
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// Call the database function to get all commits for the model
	commits, err := database.GetCommitsByModelUUID(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return the commits if found
	c.JSON(http.StatusOK, commits)
}

// GetLatestCommitByModelTag godoc
// @Summary      Get latest commit for a model
// @Description  Get latest commit for a model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Latest commit"
// @Failure      404   {object}  gin.H  "Model not found"
// @Router       /v0/models/commits/latest/{tag} [get]
func (h *ModelHandler) GetLatestCommitByModelTag(c *gin.Context) {
	tag := c.Param("tag")

	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	actingUserID, _ := getUserIDFromToken(c, h.engMode)
	hasAccess := checkIfUserHasAccessToModel(actingUserID, *model)
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// Call the encapsulated GetModelByUUID function from the database package
	commit, err := database.GetLatestCommitForModelUUID(model.Meta.UUID)
	if err != nil {
		// If error, return an appropriate response based on the error
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}

	// Return the commit if found
	c.JSON(http.StatusOK, commit)
}

// GetModelPrivacy godoc
// @Summary      Get privacy settings for model
// @Description  Get privacy settings for model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Privacy settings"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/privacy/{tag} [get]
func (h *ModelHandler) GetModelPrivacy(c *gin.Context) {
	tag := c.Param("tag")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get the model from database
	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	resp := gin.H{
		"isPublic": model.Addons.IsPublic,
		"shares":   model.Addons.Shares,
	}

	c.JSON(http.StatusOK, resp)
}

// PutModelPrivacy godoc
// @Summary      Update privacy settings for model
// @Description  Update privacy settings for model
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Param        privacy  body  object  true  "Privacy settings"
// @Success      200   {object}  gin.H  "Privacy settings updated"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Model not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/privacy/{tag} [put]
func (h *ModelHandler) PutModelPrivacy(c *gin.Context) {
	tag := c.Param("tag")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the model exists
	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// bind the json to a struct for shares
	var req struct {
		IsPublic bool             `json:"isPublic"`
		Shares   []apiTypes.Share `json:"shares"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// make sure the shares object is valid
	for _, share := range req.Shares {
		_, err := database.GetUserByEmail(share.Email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if (share.Level != "read" && share.Level != "write") || (req.IsPublic && share.Level == "read") {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid share level for user: %s", share.Email)})
			return
		}
	}

	// now update the model's privacy settings
	err = database.UpdateModelPrivacyByUUID(model.Meta.UUID, req.IsPublic, req.Shares)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
