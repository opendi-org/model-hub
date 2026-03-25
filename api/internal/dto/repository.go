package dto

import "time"

// ── Request DTOs ──────────────────────────────────────────────────────────────

// CreateRepositoryRequest represents the request body for UC-03 (Create Repository)
type CreateRepositoryRequest struct {
	Slug        string `json:"slug" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Visibility  string `json:"visibility" binding:"required,oneof=public private"`
}

// UpdateRepositoryRequest represents the request body for UC-06 (Update Repository)
type UpdateRepositoryRequest struct {
	Slug        string `json:"slug" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Visibility  string `json:"visibility" binding:"required,oneof=public private"`
}

// SearchRepositoriesQuery represents query parameters for UC-04 (Search Repositories)
type SearchRepositoriesQuery struct {
	Q     string `form:"q"`     // search text
	Scope string `form:"scope"` // "mine", "shared-with-me", "all" (default: "all")
	Owner string `form:"owner"` // when listing by owner
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// RepositoryTagInfo represents tag metadata in repository responses
type RepositoryTagInfo struct {
	Name        string    `json:"name"`
	Digest      string    `json:"digest"` // ModelUUID
	Size        int64     `json:"size"`   // SizeBytes
	LastUpdated time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"` // username
}

// CollaboratorInfo represents a collaborator in repository responses
type CollaboratorInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"` // "read", "write", "owner"
}

// RepositoryLineageInfo represents parent/child references
type RepositoryLineageInfo struct {
	Parent   *RepositoryLineageRef  `json:"parent,omitempty"`
	Children []RepositoryLineageRef `json:"children,omitempty"`
}

// RepositoryLineageRef is a minimal repo reference for lineage
type RepositoryLineageRef struct {
	ID    uint   `json:"id"`
	Owner string `json:"owner"`
	Slug  string `json:"slug"`
}

// RepositoryResponse is the standard response for repository details (UC-05)
type RepositoryResponse struct {
	ID            uint                  `json:"id"`
	Owner         string                `json:"owner"`
	Slug          string                `json:"slug"`
	Description   string                `json:"description"`
	Visibility    string                `json:"visibility"` // "public" or "private"
	CreatedAt     time.Time             `json:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
	Tags          []RepositoryTagInfo   `json:"tags"`
	Collaborators []CollaboratorInfo    `json:"collaborators,omitempty"`
	Lineage       RepositoryLineageInfo `json:"lineage,omitempty"`
}

// RepositoryListItem is a minimal response for list endpoints (UC-03, UC-04)
type RepositoryListItem struct {
	ID          uint      `json:"id"`
	Owner       string    `json:"owner"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"` // "public" or "private"
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ListRepositoriesResponse wraps a list of repositories
type ListRepositoriesResponse struct {
	Repositories []RepositoryListItem `json:"repositories"`
	Total        int64                `json:"total"`
}
