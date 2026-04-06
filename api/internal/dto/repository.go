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

// SearchRepositoriesQuery represents query parameters for both search endpoints
// /v0/repositories (auth required) and /v0/search (optional auth)
// Filters: visibility, search text, owner, sorting, scope
type SearchRepositoriesQuery struct {
	Q          string `form:"q"`          // search text (slug + description)
	Visibility string `form:"visibility"` // "public", "private", or empty for all
	Owner      string `form:"owner"`      // filter by owner username
	SortBy     string `form:"sortBy"`     // "name", "updated", "created" (default: "updated")
	SortOrder  string `form:"sortOrder"`  // "asc", "desc" (default: "desc")
	Scope      string `form:"scope"`      // "mine", "shared-with-me", "all" (default: "all")
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

// ── UC 08: Fork Repository ────────────────────────────────────────────────────

// ForkRepositoryRequest represents the request body for UC-08 (Fork Repository)
type ForkRepositoryRequest struct {
	Slug        string   `json:"slug" binding:"required,min=1,max=255"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"` // tags to copy from parent repo
}

// ForkRepositoryResponse represents the response for UC-08
type ForkRepositoryResponse struct {
	ID          uint      `json:"id"`
	Owner       string    `json:"owner"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"createdAt"`
	ForkedFrom  *struct {
		ID    uint   `json:"id"`
		Owner string `json:"owner"`
		Slug  string `json:"slug"`
	} `json:"forkedFrom,omitempty"`
}

// ── UC 09: Set Repository Privacy ──────────────────────────────────────────────

// SetPrivacyRequest represents the request body for UC-09 (Set Repository Privacy)
type SetPrivacyRequest struct {
	Visibility string `json:"visibility" binding:"required,oneof=public private"`
}

// ── UC 10: Share Repository ───────────────────────────────────────────────────

// AddCollaboratorRequest represents the request body for UC-10 (Share Repository)
type AddCollaboratorRequest struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=read write"`
}

// UpdateCollaboratorRequest represents the request body for updating collaborator role
type UpdateCollaboratorRequest struct {
	Role string `json:"role" binding:"required,oneof=read write"`
}

// CollaboratorResponse represents a collaborator in responses
type CollaboratorResponse struct {
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListCollaboratorsResponse wraps a list of collaborators
type ListCollaboratorsResponse struct {
	Collaborators []CollaboratorResponse `json:"collaborators"`
	Owner         string                 `json:"owner"`
}

// ── UC 11: Transfer Repository Ownership ────────────────────────────────────────

// TransferRepositoryRequest represents the request body for UC-11 (Transfer Repository Ownership)
type TransferRepositoryRequest struct {
	Username string `json:"username" binding:"required"`
}

// TransferRepositoryResponse represents the response for UC-11
type TransferRepositoryResponse struct {
	ID            uint      `json:"id"`
	Owner         string    `json:"owner"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	Visibility    string    `json:"visibility"`
	UpdatedAt     time.Time `json:"updatedAt"`
	PreviousOwner string    `json:"previousOwner"`
}

// ── Error responses ───────────────────────────────────────────────────────────

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
