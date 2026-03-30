package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/hub"
	"opendi.org/model-hub/api/internal/services"
)

const (
	AuthCookieName         = "opendi_access_token"
	defaultAuthTokenTTL    = 7 * 24 * time.Hour
	dbContextKey           = "_db"
	authSecretContextKey   = "_auth_secret"
	authTokenTTLContextKey = "_auth_token_ttl"
	authUserContextKey     = "_auth_user"
	authCookieSecureKey    = "_auth_cookie_secure"
)

// AttachAuthContext stores auth dependencies into request context.
func AttachAuthContext(db *gorm.DB, jwtSecret string, tokenTTL time.Duration) gin.HandlerFunc {
	return AttachAuthContextWithCookieSecure(db, jwtSecret, tokenTTL, nil)
}

// AttachAuthContextWithCookieSecure stores auth dependencies into request context.
// When cookieSecureOverride is nil, cookie Secure behavior uses current default logic.
func AttachAuthContextWithCookieSecure(db *gorm.DB, jwtSecret string, tokenTTL time.Duration, cookieSecureOverride *bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(dbContextKey, db)
		c.Set(authSecretContextKey, jwtSecret)
		if tokenTTL <= 0 {
			tokenTTL = defaultAuthTokenTTL
		}
		c.Set(authTokenTTLContextKey, tokenTTL)
		if cookieSecureOverride != nil {
			c.Set(authCookieSecureKey, *cookieSecureOverride)
		}
		c.Next()
	}
}

// AuthenticateRequest resolves current user from bearer token or auth cookie.
func AuthenticateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAuthToken(c)
		if token == "" {
			c.Next()
			return
		}

		secret, ok := c.Get(authSecretContextKey)
		if !ok {
			c.Next()
			return
		}
		secretStr, _ := secret.(string)
		claims, err := services.VerifyAccessToken(token, secretStr)
		if err != nil {
			c.Next()
			return
		}

		dbAny, exists := c.Get(dbContextKey)
		if !exists {
			c.Next()
			return
		}
		db, ok := dbAny.(*gorm.DB)
		if !ok || db == nil {
			c.Next()
			return
		}

		var user hub.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.Next()
			return
		}
		SetCurrentUser(c, &user)
		c.Next()
	}
}

func extractAuthToken(c *gin.Context) string {
	authz := strings.TrimSpace(c.GetHeader("Authorization"))
	if len(authz) >= 7 && strings.EqualFold(authz[:7], "Bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	if cookieToken, err := c.Cookie(AuthCookieName); err == nil && cookieToken != "" {
		return cookieToken
	}
	return ""
}

func GetAuthSigningConfig(c *gin.Context) (string, time.Duration, error) {
	secretAny, exists := c.Get(authSecretContextKey)
	if !exists {
		return "", 0, errors.New("auth secret not set in context")
	}
	secret, ok := secretAny.(string)
	if !ok || secret == "" {
		return "", 0, errors.New("invalid auth secret")
	}
	ttlAny, exists := c.Get(authTokenTTLContextKey)
	if !exists {
		return secret, defaultAuthTokenTTL, nil
	}
	ttl, ok := ttlAny.(time.Duration)
	if !ok || ttl <= 0 {
		return secret, defaultAuthTokenTTL, nil
	}
	return secret, ttl, nil
}

// GetAuthDB returns the request-scoped database injected by AttachAuthContext.
func GetAuthDB(c *gin.Context) (*gorm.DB, error) {
	dbAny, exists := c.Get(dbContextKey)
	if !exists {
		return nil, errors.New("database not set in context")
	}
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		return nil, errors.New("invalid database in context")
	}
	return db, nil
}

func SetAuthCookie(c *gin.Context, token string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = defaultAuthTokenTTL
	}
	secure := resolveAuthCookieSecure(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, token, int(ttl.Seconds()), "/", "", secure, true)
}

func ClearAuthCookie(c *gin.Context) {
	secure := resolveAuthCookieSecure(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, "", -1, "/", "", secure, true)
}

// GetCurrentUser returns the authenticated user cached by AuthenticateRequest.
func GetCurrentUser(c *gin.Context) (*hub.User, error) {
	if cachedUser, exists := c.Get(authUserContextKey); exists {
		if user, ok := cachedUser.(*hub.User); ok {
			return user, nil
		}
	}
	return nil, errors.New("unauthorized")
}

// OptionalGetCurrentUser returns nil when not authenticated.
func OptionalGetCurrentUser(c *gin.Context) *hub.User {
	user, err := GetCurrentUser(c)
	if err != nil {
		return nil
	}
	return user
}

// RequireAuthentication is middleware that enforces authentication.
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

// SetCurrentUser stores authenticated user on request context.
func SetCurrentUser(c *gin.Context, user *hub.User) {
	c.Set(authUserContextKey, user)
}

// GetUserIDFromContext returns authenticated user ID, or 0 if absent.
func GetUserIDFromContext(c *gin.Context) uint {
	user := OptionalGetCurrentUser(c)
	if user != nil {
		return user.ID
	}
	return 0
}

func resolveAuthCookieSecure(c *gin.Context) bool {
	if secureAny, exists := c.Get(authCookieSecureKey); exists {
		if secure, ok := secureAny.(bool); ok {
			return secure
		}
	}
	return !strings.EqualFold(gin.Mode(), gin.DebugMode)
}
