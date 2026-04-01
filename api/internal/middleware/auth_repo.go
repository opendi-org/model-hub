package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/hub"
)

// Permission levels for repository access (owner > write > read > none).
const (
	PermissionNone  = "none"
	PermissionRead  = "read"
	PermissionWrite = "write"
	PermissionOwner = "owner"
)

const (
	repoPermissionContextKey   = "permission"
	repoCollaboratorContextKey = "isCollaborator"
)

// CheckRepositoryAccess computes permission for the current request and stores:
//   - "permission" (string)
//   - "isCollaborator" (bool)
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
		c.Set(repoPermissionContextKey, permission)
		c.Set(repoCollaboratorContextKey, isCollaborator)
		c.Next()
	}
}

// RequireRepositoryPermission enforces a minimum permission level.
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
				// Conceal repository existence when caller has no visibility.
				statusCode = http.StatusNotFound
			}
			c.JSON(statusCode, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetRepositoryPermission(c *gin.Context) string {
	perm, exists := c.Get(repoPermissionContextKey)
	if !exists {
		return ""
	}
	if p, ok := perm.(string); ok {
		return p
	}
	return ""
}

func IsRepositoryCollaborator(c *gin.Context) bool {
	isCollab, exists := c.Get(repoCollaboratorContextKey)
	if !exists {
		return false
	}
	if isc, ok := isCollab.(bool); ok {
		return isc
	}
	return false
}

func getRepositoryPermissionWithCollaboratorStatus(db *gorm.DB, repo *hub.Repository, userID uint) (permission string, isCollaborator bool) {
	if userID != 0 && repo.OwnerID == userID {
		return PermissionOwner, false
	}
	if userID == 0 {
		if repo.Visibility == "public" {
			return PermissionRead, false
		}
		return PermissionNone, false
	}

	var collab hub.Collaborator
	err := db.Where("repo_id = ? AND user_id = ?", repo.ID, userID).First(&collab).Error
	if err == nil {
		switch collab.Role {
		case hub.CollaboratorRoleOwner:
			return PermissionOwner, true
		case hub.CollaboratorRoleWrite:
			return PermissionWrite, true
		case hub.CollaboratorRoleRead:
			return PermissionRead, true
		}
	}
	if repo.Visibility == "public" {
		return PermissionRead, false
	}
	return PermissionNone, false
}

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
