package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/hub"
)

// ── Authentication & Authorization ────────────────────────────────────────────
//
// This middleware provides two layers of access control:
//
// 1. AUTHENTICATION (user identity)
//    - GetCurrentUser(): Extracts and validates user from request credentials
//    - SetCurrentUser(): Caches authenticated user in request context
//    - RequireAuthentication(): Middleware enforcing authenticated requests
//
// 2. AUTHORIZATION (repository resource access)
//    - Permission constants define access levels: owner > write > read > none
//    - CheckRepositoryAccess(): Evaluates user's permission on a specific repository
//    - RequireRepositoryPermission(): Middleware enforcing permission requirements
//
// Example flow for a protected endpoint:
//   GET /v0/repositories/:owner/:slug
//     -> ResolveRepositoryByOwnerSlug middleware (fetches repo from DB)
//     -> CheckRepositoryAccess middleware (evaluates permission: read, owner, etc.)
//     -> Handler receives repo and permission level in context
//
// ──────────────────────────────────────────────────────────────────────────────

// TODO: Implement actual authentication logic here
// GetCurrentUser should extract the authenticated user from the request context
// by validating credentials (e.g., JWT token from Authorization header).
//
// Implementation guidance:
//  1. Extract token from Authorization header (Bearer scheme)
//  2. Validate token signature and expiry
//  3. Extract user claims (typically user ID or email)
//  4. Query database for hub.User record
//  5. Cache in context using SetCurrentUser() for subsequent access
//  6. Return error if token invalid, expired, or user not found
//
// Currently returns a placeholder error
func GetCurrentUser(c *gin.Context) (*hub.User, error) {
	// Check if user already cached in context (set by authentication middleware)
	if cachedUser, exists := c.Get("_auth_user"); exists {
		if user, ok := cachedUser.(*hub.User); ok {
			return user, nil
		}
	}
	return nil, errors.New("authentication not yet implemented")
}

// OptionalGetCurrentUser is like GetCurrentUser but returns nil (not an error)
// if the user is not authenticated. Used for endpoints that allow both
// authenticated and unauthenticated access (e.g., viewing public repositories).
func OptionalGetCurrentUser(c *gin.Context) *hub.User {
	user, err := GetCurrentUser(c)
	if err != nil {
		return nil
	}
	return user
}

// RequireAuthentication is middleware that enforces authentication.
// Responds with 401 Unauthorized if not authenticated.
func RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := GetCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Permission levels for a repository resource (in ascending privilege order)
//
// Permission hierarchy (each level inherits permissions of lower levels):
//
//	owner (5)  - Repository owner (OwnerID in DB) or collaborator with "owner" role
//	            Permissions: read, write, delete, transfer, manage collaborators
//
//	write (3)  - Explicit collaborator with "write" role
//	            Permissions: read, write (push/delete tags)
//
//	read  (2)  - Explicit collaborator with "read" role OR public repository
//	            Permissions: read (fetch tags and model content)
//
//	none  (0)  - No access (private repo + not a collaborator)
//	            Permissions: none
const (
	PermissionNone  = "none"  // No access
	PermissionRead  = "read"  // Read tags and model content
	PermissionWrite = "write" // Push and delete tags
	PermissionOwner = "owner" // Manage collaborators, visibility, transfer, delete (repo owner or owner collaborator)
)

// CheckRepositoryAccess is middleware that determines the authenticated user's
// permission level on a repository and stores it in context.
//
// Permissions (in order):
//   - "owner"     if user is the repository owner or a collaborator with owner role
//   - "write"     if user is a collaborator with write role
//   - "read"      if user is a collaborator with read role, or repo is public
//   - "none"      otherwise
//
// Also stores in context:
//   - "isCollaborator" (bool): true if user has explicit collaborator record (not just reading public repo)
//
// The permission level is stored in context key "permission".
// Does NOT require authentication (public readability still permits "none" → "read" for public repos).
// Must be called after ResolveRepository* middleware.
func CheckRepositoryAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo := GetRepository(c)
		if repo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not set in context"})
			c.Abort()
			return
		}

		user := OptionalGetCurrentUser(c)
		userID := uint(0)
		if user != nil {
			userID = user.ID
		}

		permission, isCollaborator := getRepositoryPermissionWithCollaboratorStatus(db, repo, userID)
		c.Set("permission", permission)
		c.Set("isCollaborator", isCollaborator)
		c.Next()
	}
}

// RequireRepositoryPermission is middleware that enforces a minimum permission
// level on the repository. Must be called after CheckRepositoryAccess middleware.
//
// requiredLevel should be one of: PermissionRead, PermissionWrite, PermissionOwner
//
// Responds with 403 Forbidden if the user lacks the required permission.
func RequireRepositoryPermission(requiredLevel string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permission := GetRepositoryPermission(c)
		if permission == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission not set in context"})
			c.Abort()
			return
		}

		if !hasRequiredPermission(permission, requiredLevel) {
			statusCode := http.StatusForbidden
			if permission == PermissionNone {
				statusCode = http.StatusUnauthorized
			}
			c.JSON(statusCode, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetRepositoryPermission retrieves the permission level from context.
// Returns empty string if not found.
func GetRepositoryPermission(c *gin.Context) string {
	perm, exists := c.Get("permission")
	if !exists {
		return ""
	}
	if p, ok := perm.(string); ok {
		return p
	}
	return ""
}

// IsRepositoryCollaborator retrieves the isCollaborator flag from context.
// Returns false if not found.
func IsRepositoryCollaborator(c *gin.Context) bool {
	isCollab, exists := c.Get("isCollaborator")
	if !exists {
		return false
	}
	if isc, ok := isCollab.(bool); ok {
		return isc
	}
	return false
}

// SetCurrentUser stores a user in the request context for later retrieval by GetCurrentUser.
// This is typically called by an authentication middleware after validating credentials.
//
// Example usage in a token validation middleware:
//
//	user, err := db.First(&user, "id = ?", userIDFromToken).Error
//	if err != nil { ... }
//	middleware.SetCurrentUser(c, user)
func SetCurrentUser(c *gin.Context, user *hub.User) {
	c.Set("_auth_user", user)
}

// GetUserIDFromContext retrieves the authenticated user's ID from context.
// Returns 0 if no authenticated user or if context key is missing.
// Useful as a convenience when you only need the user ID, not the full User record.
func GetUserIDFromContext(c *gin.Context) uint {
	user := OptionalGetCurrentUser(c)
	if user != nil {
		return user.ID
	}
	return 0
}

// ── Helper functions ──────────────────────────────────────────────────────────

// getRepositoryPermission determines the user's permission level on a repository.
// userID should be 0 for unauthenticated users.
func getRepositoryPermission(db *gorm.DB, repo *hub.Repository, userID uint) string {
	// Owner always has full access
	if userID != 0 && repo.OwnerID == userID {
		return PermissionOwner
	}

	// Unauthenticated users can only read public repos
	if userID == 0 {
		if repo.Visibility == "public" {
			return PermissionRead
		}
		return PermissionNone
	}

	// Check collaborator role
	var collab hub.Collaborator
	err := db.Where("repo_id = ? AND user_id = ?", repo.ID, userID).First(&collab).Error
	if err == nil {
		// User is an explicit collaborator
		switch collab.Role {
		case "owner":
			return PermissionOwner
		case "write":
			return PermissionWrite
		case "read":
			return PermissionRead
		}
	}

	// No stored collaborator record; check if public repo
	if repo.Visibility == "public" {
		return PermissionRead
	}

	// No access
	return PermissionNone
}

// getRepositoryPermissionWithCollaboratorStatus is like getRepositoryPermission
// but also returns whether the user is an explicit collaborator (has a collaborator record).
// This is useful for handlers that need to conditionally show internal details.
func getRepositoryPermissionWithCollaboratorStatus(db *gorm.DB, repo *hub.Repository, userID uint) (permission string, isCollaborator bool) {
	// Owner always has full access
	if userID != 0 && repo.OwnerID == userID {
		return PermissionOwner, false // Owner is not stored as a collaborator record
	}

	// Unauthenticated users can only read public repos
	if userID == 0 {
		if repo.Visibility == "public" {
			return PermissionRead, false
		}
		return PermissionNone, false
	}

	// Check collaborator role
	var collab hub.Collaborator
	err := db.Where("repo_id = ? AND user_id = ?", repo.ID, userID).First(&collab).Error
	if err == nil {
		// User is an explicit collaborator
		switch collab.Role {
		case "owner":
			return PermissionOwner, true
		case "write":
			return PermissionWrite, true
		case "read":
			return PermissionRead, true
		}
	}

	// No stored collaborator record; check if public repo
	if repo.Visibility == "public" {
		return PermissionRead, false
	}

	// No access
	return PermissionNone, false
}

// hasRequiredPermission checks if the user's current permission meets the requirement.
// Permissions are hierarchical: owner > write > read > none
func hasRequiredPermission(current, required string) bool {
	permissionHierarchy := map[string]int{
		PermissionOwner: 5,
		PermissionWrite: 3,
		PermissionRead:  2,
		PermissionNone:  0,
	}

	currentLevel, ok := permissionHierarchy[current]
	if !ok {
		return false
	}
	requiredLevel, ok := permissionHierarchy[required]
	if !ok {
		return false
	}

	return currentLevel >= requiredLevel
}
