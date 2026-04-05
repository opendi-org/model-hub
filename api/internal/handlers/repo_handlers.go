package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/models/hub"
	"opendi.org/model-hub/api/internal/services"
)

// CreateRepository handles UC-03: Create Repository
// POST /v0/repositories
// Requires: RequireAuthentication middleware.
func CreateRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req dto.CreateRepositoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := services.CreateRepository(db, user.ID, req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, services.ErrRepoAlreadyExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// ListRepositories handles UC-04: Search Repositories
// GET /v0/repositories/?q=...&scope=...&owner=...
// Authentication is optional here; public requests are supported.
func ListRepositories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query dto.SearchRepositoriesQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if query.Scope == "" {
			query.Scope = "all"
		}

		user := middleware.OptionalGetCurrentUser(c)
		params := services.ListRepositoriesParams{
			Scope:           query.Scope,
			Q:               query.Q,
			Owner:           query.Owner,
			IsAuthenticated: user != nil,
			Visibility:      query.Visibility,
			SortOrder:       query.SortOrder,
			SortBy:          query.SortBy,
		}

		if user != nil {
			params.UserID = user.ID
		}

		if (query.Scope == "mine" || query.Scope == "shared-with-me") && user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "must be authenticated to use this scope"})
			return
		}

		result, err := services.ListRepositories(db, params)
		if err != nil {
			if err.Error() == "invalid scope" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetRepository handles UC-05: View Repository
// GET /v0/repositories/:owner/:slug
// Requires: ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
func GetRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		permission := middleware.GetRepositoryPermission(c)

		includePrivate := permission == middleware.PermissionOwner || middleware.IsRepositoryCollaborator(c)
		result, err := services.GetRepositoryDetails(db, repo, includePrivate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// UpdateRepository handles UC-06: Update Repository
// PATCH /v0/repositories/:owner/:slug
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Note: RequireAuthentication depends on AuthenticateRequest at /v0 router level.
// Handler enforces PermissionOwner
func UpdateRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can update this repository"})
			return
		}

		var req dto.UpdateRepositoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := services.UpdateRepository(db, repo, req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, services.ErrRepoAlreadyExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// DeleteRepository handles UC-07: Delete Repository
// DELETE /v0/repositories/:owner/:slug
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Note: RequireAuthentication depends on AuthenticateRequest at /v0 router level.
// Handler enforces PermissionOwner only
func DeleteRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can delete this repository"})
			return
		}

		if err := services.DeleteRepository(db, repo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete repository"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

// GlobalSearch handles generic full-text search across all public repositories
// GET /v0/search?q=...&owner=...&visibility=...&sortBy=...&sortOrder=...
// No authentication required; only searches public repos unless authenticated
func GlobalSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query dto.SearchRepositoriesQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults
		if query.SortBy == "" {
			query.SortBy = "updated"
		}
		if query.SortOrder == "" {
			query.SortOrder = "desc"
		}

		user := middleware.OptionalGetCurrentUser(c)
		var userID uint
		isAuthenticated := user != nil
		if isAuthenticated {
			userID = user.ID
		}

		dbQuery := db.Model(&hub.Repository{})

		// Visibility filter: only include public repos (or user's own/shared if authenticated)
		if query.Visibility == "public" || query.Visibility == "" {
			if isAuthenticated {
				dbQuery = dbQuery.Where("visibility = 'public' OR owner_id = ? OR id IN (SELECT repo_id FROM hub_collaborators WHERE user_id = ?)", userID, userID)
			} else {
				dbQuery = dbQuery.Where("visibility = 'public'")
			}
		}

		// Full-text search on slug and description
		if query.Q != "" {
			searchTerm := "%" + strings.ToLower(query.Q) + "%"
			dbQuery = dbQuery.Where("LOWER(slug) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
		}

		// Apply owner filter
		if query.Owner != "" {
			var ownerUser hub.User
			if err := db.Model(&hub.User{}).Where("username = ?", query.Owner).First(&ownerUser).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// No repos for this owner
					c.JSON(http.StatusOK, dto.ListRepositoriesResponse{
						Repositories: []dto.RepositoryListItem{},
						Total:        0,
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			dbQuery = dbQuery.Where("owner_id = ?", ownerUser.ID)
		}

		// Get total count
		var total int64
		if err := dbQuery.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Sorting
		sortColumn := "updated_at"
		switch query.SortBy {
		case "name":
			sortColumn = "slug"
		case "created":
			sortColumn = "created_at"
		case "updated":
			sortColumn = "updated_at"
		}

		sortDir := "DESC"
		if query.SortOrder == "asc" {
			sortDir = "ASC"
		}

		// Apply sorting
		dbQuery = dbQuery.Order(sortColumn + " " + sortDir)

		// Fetch repositories with owners
		var repos []hub.Repository
		if err := dbQuery.Preload("Owner").Find(&repos).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		items := make([]dto.RepositoryListItem, len(repos))
		for i, repo := range repos {
			items[i] = dto.RepositoryListItem{
				ID:          repo.ID,
				Owner:       repo.Owner.Username,
				Slug:        repo.Slug,
				Description: repo.Description,
				Visibility:  repo.Visibility,
				CreatedAt:   repo.CreatedAt,
				UpdatedAt:   repo.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, dto.ListRepositoriesResponse{
			Repositories: items,
			Total:        total,
		})
	}
}

// Forks Repository handles UC-08: Fork Repository
// POST /v0/repositories/:owner/:slug/fork
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Handler requires read access to source repository
func ForkRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		sourceRepo := middleware.GetRepository(c)
		if sourceRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		// Check read access to source repo
		if middleware.GetRepositoryPermission(c) == middleware.PermissionNone {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient access to source repository"})
			return
		}

		var req dto.ForkRepositoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate slug format
		if !isValidSlug(req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository name: must be 1-255 characters and contain only letters, numbers, hyphens (-), and underscores (_)"})
			return
		}

		// Check for duplicate slug under this owner
		var count int64
		if err := db.Model(&hub.Repository{}).
			Where("owner_id = ? AND slug = ?", user.ID, req.Slug).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "repository name already exists"})
			return
		}

		// Create forked repository
		forkedRepo := &hub.Repository{
			OwnerID:      user.ID,
			Slug:         req.Slug,
			Description:  req.Description,
			Visibility:   "private", // Forks are always private initially
			ForkedFromID: &sourceRepo.ID,
		}

		if err := db.Create(forkedRepo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create forked repository"})
			return
		}

		// Copy selected tags from source repository
		if len(req.Tags) > 0 {
			// Get tags from source repo
			var sourceTags []hub.CDMTag
			if err := db.Where("repo_id = ? AND name IN ?", sourceRepo.ID, req.Tags).Find(&sourceTags).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to copy tags"})
				return
			}

			// Create new tags under forked repo
			for _, srcTag := range sourceTags {
				newTag := &hub.CDMTag{
					RepoID:    forkedRepo.ID,
					Name:      srcTag.Name,
					ModelUUID: srcTag.ModelUUID,
					SizeBytes: srcTag.SizeBytes,
					CreatedBy: srcTag.CreatedBy,
				}
				if err := db.Create(newTag).Error; err != nil {
					// Log but don't fail the fork operation
					continue
				}
			}
		}

		// Fetch owner to build response
		if err := db.Preload("Owner").Preload("ForkedFrom").First(forkedRepo, forkedRepo.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response := dto.ForkRepositoryResponse{
			ID:          forkedRepo.ID,
			Owner:       user.Username,
			Slug:        forkedRepo.Slug,
			Description: forkedRepo.Description,
			Visibility:  forkedRepo.Visibility,
			CreatedAt:   forkedRepo.CreatedAt,
		}

		if forkedRepo.ForkedFromID != nil {
			response.ForkedFrom = &struct {
				ID    uint   `json:"id"`
				Owner string `json:"owner"`
				Slug  string `json:"slug"`
			}{
				ID:    sourceRepo.ID,
				Owner: sourceRepo.Owner.Username,
				Slug:  sourceRepo.Slug,
			}
		}

		c.JSON(http.StatusCreated, response)
	}
}

// SetRepositoryPrivacy handles UC-09: Set Repository Privacy
// PATCH /v0/repositories/:owner/:slug (with visibility field)
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Handler requires owner permission
func SetRepositoryPrivacy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can change repository privacy"})
			return
		}

		var req dto.SetPrivacyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate visibility value
		if req.Visibility != "public" && req.Visibility != "private" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "visibility must be 'public' or 'private'"})
			return
		}

		// Update visibility
		if err := db.Model(repo).Update("visibility", req.Visibility).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update repository privacy"})
			return
		}

		// Fetch updated repo with owner
		if err := db.Preload("Owner").First(repo, repo.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response := dto.RepositoryListItem{
			ID:          repo.ID,
			Owner:       repo.Owner.Username,
			Slug:        repo.Slug,
			Description: repo.Description,
			Visibility:  repo.Visibility,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
		}

		c.JSON(http.StatusOK, response)
	}
}

// AddCollaborator handles UC-10: Add/Update Collaborator
// PUT /v0/repositories/:owner/:slug/collaborators/:username
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Handler requires owner permission
func AddCollaborator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can manage collaborators"})
			return
		}

		username := c.Param("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		var req dto.AddCollaboratorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate role
		if req.Role != "read" && req.Role != "write" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'read' or 'write'"})
			return
		}

		// Find the user to add
		var targetUser hub.User
		if err := db.Where("username = ?", username).First(&targetUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Prevent sharing with self
		if targetUser.ID == user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot share repository with yourself"})
			return
		}

		// Check for existing collaborator
		var existingCollab hub.Collaborator
		err := db.Where("repo_id = ? AND user_id = ?", repo.ID, targetUser.ID).First(&existingCollab).Error

		if err == nil {
			// Update existing collaborator
			if existingCollab.Role == req.Role {
				c.JSON(http.StatusConflict, gin.H{"error": "user already has this role"})
				return
			}
			if err := db.Model(&existingCollab).Update("role", req.Role).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update collaborator role"})
				return
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new collaborator
			newCollab := &hub.Collaborator{
				RepoID: repo.ID,
				UserID: targetUser.ID,
				Role:   req.Role,
			}
			if err := db.Create(newCollab).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add collaborator"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response := dto.CollaboratorResponse{
			Username:  targetUser.Username,
			Role:      req.Role,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// RemoveCollaborator handles UC-10: Remove Collaborator (revoke access)
// DELETE /v0/repositories/:owner/:slug/collaborators/:username
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Handler requires owner permission
func RemoveCollaborator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can manage collaborators"})
			return
		}

		username := c.Param("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		// Find the user to remove
		var targetUser hub.User
		if err := db.Where("username = ?", username).First(&targetUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Remove the collaborator
		result := db.Where("repo_id = ? AND user_id = ?", repo.ID, targetUser.ID).Delete(&hub.Collaborator{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove collaborator"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "collaborator not found"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

// ListCollaborators handles UC-10: List Collaborators
// GET /v0/repositories/:owner/:slug/collaborators
// Requires: ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Returns collaborators only to owner and explicit collaborators
func ListCollaborators(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		permission := middleware.GetRepositoryPermission(c)

		// Only show collaborators to owner and explicit collaborators
		if permission != middleware.PermissionOwner && !middleware.IsRepositoryCollaborator(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient access to view collaborators"})
			return
		}

		var collabs []hub.Collaborator
		if err := db.Where("repo_id = ?", repo.ID).Preload("User").Find(&collabs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response := dto.ListCollaboratorsResponse{
			Collaborators: make([]dto.CollaboratorResponse, len(collabs)),
			Owner:         repo.Owner.Username,
		}

		for i, collab := range collabs {
			response.Collaborators[i] = dto.CollaboratorResponse{
				Username:  collab.User.Username,
				Role:      collab.Role,
				CreatedAt: collab.CreatedAt,
				UpdatedAt: collab.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// TransferRepositoryOwnership handles UC-11: Transfer Repository Ownership
// POST /v0/repositories/:owner/:slug/transfer
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Handler requires owner permission
func TransferRepositoryOwnership(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can transfer the repository"})
			return
		}

		var req dto.TransferRepositoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		// Check if username is self-transfer
		if req.Username == user.Username {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer repository to yourself"})
			return
		}

		// Find the new owner
		var newOwner hub.User
		if err := db.Where("username = ?", req.Username).First(&newOwner).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Store previous owner for response
		previousOwner := user.Username

		// Perform transfer within a transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			// Update owner explicitly with WHERE clause
			if result := tx.Model(&hub.Repository{}).Where("id = ?", repo.ID).Update("owner_id", newOwner.ID); result.Error != nil {
				return result.Error
			}

			// Remove new owner from collaborators if they exist
			if err := tx.Where("repo_id = ? AND user_id = ?", repo.ID, newOwner.ID).Delete(&hub.Collaborator{}).Error; err != nil {
				return err
			}

			// Optionally: add previous owner as owner collaborator
			// (The requirements document doesn't specify this behavior, so we skip it for now)

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transfer repository"})
			return
		}

		// Fetch updated repo with new owner (use fresh struct to ensure proper reload)
		var updatedRepo hub.Repository
		if err := db.Preload("Owner").First(&updatedRepo, repo.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response := dto.TransferRepositoryResponse{
			ID:            updatedRepo.ID,
			Owner:         updatedRepo.Owner.Username,
			Slug:          updatedRepo.Slug,
			Description:   updatedRepo.Description,
			Visibility:    updatedRepo.Visibility,
			UpdatedAt:     updatedRepo.UpdatedAt,
			PreviousOwner: previousOwner,
		}

		c.JSON(http.StatusOK, response)
	}
}

// PutTagModel handles UC-13: Upload Model
// PUT /v0/repositories/:owner/:slug/tags/:tag
// Requires: ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
func PutTagModel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if !middleware.HasRequiredPermission(middleware.GetRepositoryPermission(c), middleware.PermissionWrite) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		tagName := c.Param("tag")
		if !isValidTagName(tagName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag name: only alphanumeric, hyphens, underscores, and dots allowed"})
			return
		}

		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		if len(raw) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
			return
		}

		user, err := middleware.GetCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		result, err := services.UploadModel(db, repo.ID, tagName, raw, user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tag":    result.Tag,
			"digest": result.Digest,
			"size":   result.Size,
		})
	}
}

// GetTagModel handles UC-12: Download Model
// GET /v0/repositories/:owner/:slug/tags/:tag/model
// Requires: ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
func GetTagModel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) == middleware.PermissionNone {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		result, err := services.DownloadModel(db, repo.ID, c.Param("tag"))
		if err != nil {
			if errors.Is(err, services.ErrTagNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load model"})
			return
		}

		c.Header("ETag", result.Digest)
		if c.GetHeader("If-None-Match") == result.Digest {
			c.Status(http.StatusNotModified)
			return
		}

		c.JSON(http.StatusOK, result.Model)
	}
}

// ── Helper functions ──────────────────────────────────────────────────────────

// isValidTagName checks if the tag name is valid. Tags allow alphanumeric,
// hyphens, underscores, and dots (e.g. "v1.0", "latest", "my-tag").
func isValidTagName(tag string) bool {
	if len(tag) == 0 || len(tag) > 255 {
		return false
	}
	for _, ch := range tag {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.') {
			return false
		}
	}
	return true
}

// isValidSlug checks that a slug is alphanumeric plus hyphens/underscores, 1-255 chars.
func isValidSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 255 {
		return false
	}
	for _, ch := range slug {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}
