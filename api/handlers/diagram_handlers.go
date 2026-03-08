package handlers

import (
	"net/http"
	"opendi/model-hub/api/apiTypes"

	"github.com/gin-gonic/gin"
)

// DiagramHandler serves endpoints that expose a simplified graph of repos,
// models, and tags for frontend visualization.
type DiagramHandler struct{}

func NewDiagramHandler() *DiagramHandler {
	return &DiagramHandler{}
}

// GetReposForUser returns the repos associated with a given diagram user.
func (h *DiagramHandler) GetReposForUser(c *gin.Context) {
	userID := c.Param("userID")

	repos, ok := apiTypes.GetReposForUser(userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId": userID,
		"repos":  apiTypes.ToRepoSummaries(repos),
	})
}

// GetModelsForRepo returns the models (and their versions/tags) for a repo.
func (h *DiagramHandler) GetModelsForRepo(c *gin.Context) {
	repoID := c.Param("repoID")

	models, ok := apiTypes.GetModelsForRepo(repoID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}

	c.JSON(http.StatusOK, apiTypes.ToModelSummaries(models))
}

// GetDiagramForRepo returns a graph-style payload with nodes and edges that
// mirror the whiteboard diagram for a particular repo.
func (h *DiagramHandler) GetDiagramForRepo(c *gin.Context) {
	repoID := c.Param("repoID")

	payload, ok := apiTypes.BuildDiagramForRepo(repoID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}

	c.JSON(http.StatusOK, payload)
}

