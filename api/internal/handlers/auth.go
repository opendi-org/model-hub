package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/models/hub"
	"opendi.org/model-hub/api/internal/services"
)

// AuthMe returns profile data for the authenticated user.
// Requires RequireAuthentication middleware upstream.
func AuthMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := middleware.GetCurrentUser(c)
		c.JSON(http.StatusOK, dto.MeResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
		})
	}
}

// AuthLogout clears the auth cookie for web clients.
// JWTs are stateless in MVP, so bearer tokens are not server-revoked yet.
func AuthLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware.ClearAuthCookie(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// GoogleStart redirects to Google OAuth.
// If cli_code is provided, it is sent through OAuth state for CLI approval flow.
func GoogleStart(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cliCode := strings.TrimSpace(c.Query("cli_code"))
		username := strings.TrimSpace(c.Query("username"))
		if cliCode != "" {
			if ok := tryApproveCLIWithExistingSession(c, cliCode, cfg.GoogleRedirectURL); ok {
				return
			}
		}

		mode := "web"
		if cliCode != "" {
			mode = "cli"
		}
		state, err := services.BuildOAuthState(cfg.JWTSecret, mode, cliCode, username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create oauth state"})
			return
		}
		url := services.BuildGoogleAuthorizeURL(cfg.GoogleClientID, cfg.GoogleRedirectURL, state)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// GoogleCallback resolves Google identity, creates/loads local user, and issues access token.
// For first login, username must be provided as query param.
func GoogleCallback(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authSvc := services.NewAuthService(db)
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}
		stateToken := c.Query("state")
		state, err := services.ParseOAuthState(stateToken, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
			return
		}

		cliErr := func(status int, msg string) {
			if state.Mode == "cli" && state.CLICode != "" {
				c.JSON(status, gin.H{"error": msg, "cli_code": state.CLICode})
				return
			}
			c.JSON(status, gin.H{"error": msg})
		}

		identity, err := services.ExchangeGoogleCode(c.Request.Context(), code, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "google login failed"})
			return
		}

		username := strings.TrimSpace(state.Username)
		if username == "" {
			username = c.Query("username")
		}

		var user *hub.User
		// If username is provided, it's a signup flow
		if username != "" {
			var err error
			user, err = authSvc.ResolveOrCreateUser(&dto.GoogleIdentity{
				Sub:   identity.Sub,
				Email: identity.Email,
				Name:  identity.Name,
			}, username)
			if err != nil {
				if errors.Is(err, services.ErrUsernameRequired) || errors.Is(err, services.ErrInvalidUsername) {
					cliErr(http.StatusBadRequest, err.Error())
					return
				}
				if errors.Is(err, services.ErrUsernameAlreadyTaken) {
					cliErr(http.StatusConflict, err.Error())
					return
				}
				if errors.Is(err, services.ErrAccountAlreadyExists) {
					cliErr(http.StatusConflict, err.Error())
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user"})
				return
			}
		} else {
			// No username provided, it's a signin flow
			var err error
			user, err = authSvc.ResolveExistingUser(&dto.GoogleIdentity{
				Sub:   identity.Sub,
				Email: identity.Email,
				Name:  identity.Name,
			})
			if err != nil {
				if errors.Is(err, services.ErrAccountNotFound) {
					cliErr(http.StatusBadRequest, err.Error())
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user"})
				return
			}
		}

		if state.Mode == "cli" && state.CLICode != "" {
			if err := authSvc.ApproveCLISession(state.CLICode, user.ID); err != nil {
				if errors.Is(err, services.ErrCLIInvalidOrExpired) {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve cli session"})
				return
			}
		}

		secret, ttl, err := middleware.GetAuthSigningConfig(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth config missing"})
			return
		}
		token, err := services.IssueAccessToken(user.ID, user.Username, secret, ttl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}
		middleware.SetAuthCookie(c, token, ttl)
		accessToken := ""
		if state.Mode == "cli" {
			accessToken = token
		}
		c.JSON(http.StatusOK, dto.TokenResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   int64(ttl.Seconds()),
			CliCode: func() string {
				if state.Mode == "cli" {
					return state.CLICode
				}
				return ""
			}(),
		})
	}
}

// CLILogin starts pre-auth approval login for CLI clients.
func CLILogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authSvc := services.NewAuthService(db)
		code, err := randomHex(48)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create code"})
			return
		}
		expiresAt := time.Now().UTC().Add(10 * time.Minute)

		if err := authSvc.CreateCLISession(code, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cli session"})
			return
		}

		c.JSON(http.StatusOK, dto.CLILoginResponse{
			Code:      code,
			LoginURL:  "/v0/auth/login/google/start?cli_code=" + code,
			ExpiresIn: int64(time.Until(expiresAt).Seconds()),
		})
	}
}

// CLIPoll checks approval status and returns bearer token when approved.
func CLIPoll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authSvc := services.NewAuthService(db)
		var req dto.CLIPollRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cliUser, err := authSvc.ConsumeApprovedCLISession(req.Code)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrCLICodeExpired):
				c.JSON(http.StatusGone, gin.H{"error": "code expired"})
				return
			case errors.Is(err, services.ErrCLIPending):
				c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
				return
			case errors.Is(err, services.ErrCLINotApprovable):
				c.JSON(http.StatusConflict, gin.H{"error": "cli session is not approvable"})
				return
			case errors.Is(err, services.ErrCLINotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "unknown code"})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
		}

		secret, ttl, err := middleware.GetAuthSigningConfig(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth config missing"})
			return
		}
		token, err := services.IssueAccessToken(cliUser.ID, cliUser.Username, secret, ttl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}
		c.JSON(http.StatusOK, dto.TokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(ttl.Seconds()),
		})
	}
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf)[:n], nil
}

func tryApproveCLIWithExistingSession(c *gin.Context, cliCode string, googleRedirectURL string) bool {
	user, _ := middleware.GetCurrentUser(c)
	if user == nil {
		return false
	}
	db, err := middleware.GetAuthDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not set in context"})
		return true
	}
	authSvc := services.NewAuthService(db)
	if err := authSvc.ApproveCLISession(cliCode, user.ID); err != nil {
		if errors.Is(err, services.ErrCLIInvalidOrExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return true
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve cli session"})
		return true
	}
	c.Redirect(http.StatusTemporaryRedirect, frontendCLIApprovedURL(googleRedirectURL))
	return true
}

func frontendCLIApprovedURL(googleRedirectURL string) string {
	u, err := url.Parse(strings.TrimSpace(googleRedirectURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "/auth/cli-approved"
	}

	p := path.Clean(u.Path)
	if strings.HasSuffix(p, "/auth/callback") {
		p = strings.TrimSuffix(p, "/auth/callback")
	} else {
		p = path.Dir(p)
	}
	if p == "." || p == "/" {
		u.Path = "/auth/cli-approved"
	} else {
		u.Path = path.Join(p, "auth", "cli-approved")
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
