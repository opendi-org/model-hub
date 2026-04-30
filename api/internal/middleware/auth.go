package middleware

import (
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

// Authentication extraction/verification is implemented in auth_context.go.
// This file contains authorization helpers and permission hierarchy checks.


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
		case "admin":
			return PermissionAdmin
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

// HasRequiredPermission checks if the user's current permission meets the requirement.
// Permissions are hierarchical: owner > admin > write > read > none
func HasRequiredPermission(current, required string) bool {
	permissionHierarchy := map[string]int{
		PermissionOwner: 5,
		PermissionAdmin: 4,
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
