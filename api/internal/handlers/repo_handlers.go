package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/models/hub"
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

		// Validate slug format: alphanumeric, -, _, max 255 chars
		if !isValidSlug(req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository name"})
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

		// Create the repository
		repo := &hub.Repository{
			OwnerID:     user.ID,
			Slug:        req.Slug,
			Description: req.Description,
			Visibility:  req.Visibility,
		}

		if err := db.Create(repo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create repository"})
			return
		}

		// Fetch owner user to get username
		var ownerUser hub.User
		if err := db.Model(&hub.User{}).Where("id = ?", user.ID).First(&ownerUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
			return
		}

		response := dto.RepositoryListItem{
			ID:          repo.ID,
			Owner:       ownerUser.Username,
			Slug:        repo.Slug,
			Description: repo.Description,
			Visibility:  repo.Visibility,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
		}

		c.JSON(http.StatusCreated, response)
	}
}

// ListRepositories handles UC-04: Search Repositories
// GET /v0/repositories/?q=...&owner=...&visibility=...&sortBy=...&sortOrder=...
// Search across user's repositories (owned or shared)
func ListRepositories(db *gorm.DB) gin.HandlerFunc {
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

		// Get authenticated user (required by middleware)
		user, err := middleware.GetCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		dbQuery := db.Model(&hub.Repository{})

		// Show all accessible repositories: owned + shared (for personal search)
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT repo_id FROM hub_collaborators WHERE user_id = ?)", user.ID, user.ID)

		// Apply visibility filter
		if query.Visibility != "" && (query.Visibility == "public" || query.Visibility == "private") {
			dbQuery = dbQuery.Where("visibility = ?", query.Visibility)
		}

		// Apply search filter
		if query.Q != "" {
			searchTerm := "%" + strings.ToLower(query.Q) + "%"
			dbQuery = dbQuery.Where("LOWER(slug) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
		}

		// Apply owner filter
		if query.Owner != "" {
			ownerName := strings.ToLower(strings.TrimSpace(query.Owner))
			var ownerUser hub.User
			if err := db.Model(&hub.User{}).Where("username = ?", ownerName).First(&ownerUser).Error; err != nil {
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

		// Apply sorting
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

		// Fetch repositories
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

// GetRepository handles UC-05: View Repository
// GET /v0/repositories/:owner/:slug
// Requires: ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware (no authentication required)
// Handler enforces read access (permission != PermissionNone)
func GetRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Repository and permission already resolved by middleware
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		permission := middleware.GetRepositoryPermission(c)

		// Build response
		response := dto.RepositoryResponse{
			ID:          repo.ID,
			Owner:       repo.Owner.Username,
			Slug:        repo.Slug,
			Description: repo.Description,
			Visibility:  repo.Visibility,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
		}

		// Fetch tags
		var tags []hub.CDMTag
		if err := db.Where("repo_id = ?", repo.ID).Preload("CreatedBy").Find(&tags).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		response.Tags = make([]dto.RepositoryTagInfo, len(tags))
		for i, tag := range tags {
			response.Tags[i] = dto.RepositoryTagInfo{
				Name:        tag.Name,
				Digest:      tag.ModelUUID,
				Size:        tag.SizeBytes,
				LastUpdated: tag.UpdatedAt,
				CreatedBy:   tag.CreatedBy.Username,
			}
		}

		// Fetch collaborators (only show to owner and explicit collaborators)
		isOwner := permission == middleware.PermissionOwner
		isExplicitCollaborator := middleware.IsRepositoryCollaborator(c)
		if isOwner || isExplicitCollaborator {
			var collabs []hub.Collaborator
			if err := db.Where("repo_id = ?", repo.ID).Preload("User").Find(&collabs).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			response.Collaborators = make([]dto.CollaboratorInfo, len(collabs))
			for i, collab := range collabs {
				response.Collaborators[i] = dto.CollaboratorInfo{
					Username: collab.User.Username,
					Role:     collab.Role,
				}
			}
		}

		// Fetch lineage if accessible (owner or explicit collab only)
		if isOwner || isExplicitCollaborator {
			lineage := dto.RepositoryLineageInfo{}

			// Parent
			if repo.ForkedFromID != nil {
				var parent hub.Repository
				if err := db.Unscoped().Preload("Owner").Where("id = ?", *repo.ForkedFromID).First(&parent).Error; err == nil {
					lineage.Parent = &dto.RepositoryLineageRef{
						ID:    parent.ID,
						Owner: parent.Owner.Username,
						Slug:  parent.Slug,
					}
				}
			}

			// Children
			var children []hub.Repository
			if err := db.Where("forked_from_id = ? AND deleted_at IS NULL", repo.ID).Preload("Owner").Find(&children).Error; err == nil {
				lineage.Children = make([]dto.RepositoryLineageRef, len(children))
				for i, child := range children {
					lineage.Children[i] = dto.RepositoryLineageRef{
						ID:    child.ID,
						Owner: child.Owner.Username,
						Slug:  child.Slug,
					}
				}
			}

			response.Lineage = lineage
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateRepository handles UC-06: Update Repository
// PATCH /v0/repositories/:owner/:slug
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Note: RequireAuthentication depends on AuthenticateRequest at /v0 router level.
// Handler enforces PermissionOwner
func UpdateRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Repository and permission already resolved by middleware
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

		// Validate new slug
		if !isValidSlug(req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository name"})
			return
		}

		// Check for duplicate slug if changing the name
		if req.Slug != repo.Slug {
			var count int64
			if err := db.Model(&hub.Repository{}).
				Where("owner_id = ? AND slug = ?", repo.OwnerID, req.Slug).
				Count(&count).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "repository name already exists"})
				return
			}
		}

		// Update repository
		updateData := map[string]interface{}{
			"slug":        req.Slug,
			"description": req.Description,
			"visibility":  req.Visibility,
		}

		if err := db.Model(repo).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update repository"})
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

// DeleteRepository handles UC-07: Delete Repository
// DELETE /v0/repositories/:owner/:slug
// Requires: RequireAuthentication, ResolveRepositoryByOwnerSlug, CheckRepositoryAccess middleware
// Note: RequireAuthentication depends on AuthenticateRequest at /v0 router level.
// Handler enforces PermissionOwner only
func DeleteRepository(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Repository and permission already resolved by middleware
		repo := middleware.GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			return
		}

		if middleware.GetRepositoryPermission(c) != middleware.PermissionOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can delete this repository"})
			return
		}

		// Soft-delete the repository
		if err := db.Delete(repo).Error; err != nil {
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

// ── Helper functions ──────────────────────────────────────────────────────────

// isValidSlug checks if the slug is valid for use as a repository name.
// Slugs must be alphanumeric, hyphens, and underscores, 1-255 chars.
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
