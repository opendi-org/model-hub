package services

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/models/hub"
)

// Sentinel errors for repository operations.
var (
	ErrRepoNotFound      = errors.New("repository not found")
	ErrRepoAlreadyExists = errors.New("repository name already exists")
	ErrInvalidSlug       = errors.New("invalid repository name")
)

// CreateRepository validates and persists a new repository owned by ownerID.
func CreateRepository(db *gorm.DB, ownerID uint, req dto.CreateRepositoryRequest) (*dto.RepositoryListItem, error) {
	if !isValidSlug(req.Slug) {
		return nil, ErrInvalidSlug
	}

	var count int64
	if err := db.Model(&hub.Repository{}).
		Where("owner_id = ? AND slug = ?", ownerID, req.Slug).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrRepoAlreadyExists
	}

	repo := &hub.Repository{
		OwnerID:     ownerID,
		Slug:        req.Slug,
		Description: req.Description,
		Visibility:  req.Visibility,
	}
	if err := db.Create(repo).Error; err != nil {
		return nil, err
	}

	var owner hub.User
	if err := db.Where("id = ?", ownerID).First(&owner).Error; err != nil {
		return nil, err
	}

	return &dto.RepositoryListItem{
		ID:          repo.ID,
		Owner:       owner.Username,
		Slug:        repo.Slug,
		Description: repo.Description,
		Visibility:  repo.Visibility,
		CreatedAt:   repo.CreatedAt,
		UpdatedAt:   repo.UpdatedAt,
	}, nil
}

// ListRepositoriesParams carries the inputs for listing/searching repositories.
type ListRepositoriesParams struct {
	Scope           string
	Q               string
	Owner           string
	IsAuthenticated bool
	UserID          uint
	Visibility      string
	SortOrder       string
	SortBy          string
}

// ListRepositories queries repositories according to scope, search, and owner filters.
func ListRepositories(db *gorm.DB, params ListRepositoriesParams) (*dto.ListRepositoriesResponse, error) {
	// Set defaults
	if params.SortBy == "" {
		params.SortBy = "updated"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	dbQuery := db.Model(&hub.Repository{})

	switch params.Scope {
	case "mine":
		dbQuery = dbQuery.Where("owner_id = ?", params.UserID)
	case "shared-with-me":
		// Only show repositories shared with user (collaborator)
		dbQuery = dbQuery.Where("id IN (SELECT repo_id FROM hub_collaborators WHERE user_id = ?)", params.UserID)
	case "all":
		if params.IsAuthenticated {
			dbQuery = dbQuery.Where(
				"visibility = 'public' OR owner_id = ? OR id IN (SELECT repo_id FROM hub_collaborators WHERE user_id = ?)",
				params.UserID, params.UserID,
			)
		} else {
			dbQuery = dbQuery.Where("visibility = 'public'")
		}
	default:
		return nil, errors.New("invalid scope")
	}

	if params.Visibility != "" && (params.Visibility == "public" || params.Visibility == "private") {
		dbQuery = dbQuery.Where("visibility = ?", params.Visibility)
	}

	if params.Q != "" {
		term := "%" + strings.ToLower(params.Q) + "%"
		dbQuery = dbQuery.Where("LOWER(slug) LIKE ? OR LOWER(description) LIKE ?", term, term)
	}

	if params.Owner != "" {
		ownerName := strings.ToLower(strings.TrimSpace(params.Owner))
		var ownerUser hub.User
		if err := db.Where("username = ?", ownerName).First(&ownerUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &dto.ListRepositoriesResponse{Repositories: []dto.RepositoryListItem{}, Total: 0}, nil
			}
			return nil, err
		}
		dbQuery = dbQuery.Where("owner_id = ?", ownerUser.ID)
	}

	// Apply sorting
	sortColumn := "updated_at"
	switch params.SortBy {
	case "name":
		sortColumn = "slug"
	case "created":
		sortColumn = "created_at"
	case "updated":
		sortColumn = "updated_at"
	}

	sortDir := "DESC"
	if params.SortOrder == "asc" {
		sortDir = "ASC"
	}

	// Apply sorting
	dbQuery = dbQuery.Order(sortColumn + " " + sortDir)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var repos []hub.Repository
	if err := dbQuery.Preload("Owner").Find(&repos).Error; err != nil {
		return nil, err
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

	return &dto.ListRepositoriesResponse{Repositories: items, Total: total}, nil
}

// GetRepositoryDetails fetches tags, collaborators, and lineage for a repository.
// includePrivate controls whether collaborators and lineage are populated.
func GetRepositoryDetails(db *gorm.DB, repo *hub.Repository, includePrivate bool) (*dto.RepositoryResponse, error) {
	response := &dto.RepositoryResponse{
		ID:          repo.ID,
		Owner:       repo.Owner.Username,
		Slug:        repo.Slug,
		Description: repo.Description,
		Visibility:  repo.Visibility,
		CreatedAt:   repo.CreatedAt,
		UpdatedAt:   repo.UpdatedAt,
	}

	var tags []hub.CDMTag
	if err := db.Where("repo_id = ?", repo.ID).Preload("CreatedBy").Find(&tags).Error; err != nil {
		return nil, err
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

	if includePrivate {
		var collabs []hub.Collaborator
		if err := db.Where("repo_id = ?", repo.ID).Preload("User").Find(&collabs).Error; err != nil {
			return nil, err
		}
		response.Collaborators = make([]dto.CollaboratorInfo, len(collabs))
		for i, collab := range collabs {
			response.Collaborators[i] = dto.CollaboratorInfo{
				Username: collab.User.Username,
				Role:     collab.Role,
			}
		}

		lineage := dto.RepositoryLineageInfo{}
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

	return response, nil
}

// UpdateRepository applies slug/description/visibility changes to a repository.
func UpdateRepository(db *gorm.DB, repo *hub.Repository, req dto.UpdateRepositoryRequest) (*dto.RepositoryListItem, error) {
	if !isValidSlug(req.Slug) {
		return nil, ErrInvalidSlug
	}

	if req.Slug != repo.Slug {
		var count int64
		if err := db.Model(&hub.Repository{}).
			Where("owner_id = ? AND slug = ?", repo.OwnerID, req.Slug).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrRepoAlreadyExists
		}
	}

	if err := db.Model(repo).Updates(map[string]interface{}{
		"slug":        req.Slug,
		"description": req.Description,
		"visibility":  req.Visibility,
	}).Error; err != nil {
		return nil, err
	}

	if err := db.Preload("Owner").First(repo, repo.ID).Error; err != nil {
		return nil, err
	}

	return &dto.RepositoryListItem{
		ID:          repo.ID,
		Owner:       repo.Owner.Username,
		Slug:        repo.Slug,
		Description: repo.Description,
		Visibility:  repo.Visibility,
		CreatedAt:   repo.CreatedAt,
		UpdatedAt:   repo.UpdatedAt,
	}, nil
}

// DeleteRepository soft-deletes the repository, but first mutates the slug so it
// can be reused immediately. We can't hard-delete because other tables (e.g.
// tags) have foreign keys pointing at repositories.
func DeleteRepository(db *gorm.DB, repo *hub.Repository) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Keep within the API validation limit (<=255).
		suffix := fmt.Sprintf("__deleted__%d", repo.ID)
		base := repo.Slug
		maxLen := 255
		baseMax := maxLen - len(suffix)
		if baseMax < 1 {
			baseMax = 1
		}
		if len(base) > baseMax {
			base = base[:baseMax]
		}
		newSlug := base + suffix

		if err := tx.Model(repo).Update("slug", newSlug).Error; err != nil {
			return err
		}
		return tx.Delete(repo).Error
	})
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
