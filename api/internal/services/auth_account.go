package services

import (
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/models/hub"
)

var (
	ErrUsernameRequired     = errors.New("username required for first login")
	ErrInvalidUsername      = errors.New("invalid username")
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrAccountAlreadyExists = errors.New("account already exists; please sign in instead")
	ErrAccountNotFound      = errors.New("account not found; please sign up first")
	ErrCLIInvalidOrExpired  = errors.New("invalid or expired cli approval code")
	ErrCLINotFound          = errors.New("unknown code")
	ErrCLIPending           = errors.New("pending")
	ErrCLINotApprovable     = errors.New("cli session is not approvable")
	ErrCLICodeExpired       = errors.New("code expired")
	errOAuthIdentityExists  = errors.New("oauth identity already exists")
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) ApproveCLISession(code string, userID uint) error {
	code = strings.TrimSpace(code)
	if code == "" || userID == 0 {
		return ErrCLIInvalidOrExpired
	}
	tx := s.db.Model(&hub.CLISession{}).
		Where("code = ? AND status = ? AND expires_at > ?", code, "pending", time.Now().UTC()).
		Updates(map[string]any{
			"status":  "approved",
			"user_id": userID,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrCLIInvalidOrExpired
	}
	return nil
}

func (s *AuthService) CreateCLISession(code string, expiresAt time.Time) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrCLIInvalidOrExpired
	}
	session := &hub.CLISession{
		Code:      code,
		Status:    "pending",
		ExpiresAt: expiresAt,
	}
	return s.db.Create(session).Error
}

func (s *AuthService) ConsumeApprovedCLISession(code string) (*hub.User, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrCLINotFound
	}
	var outUser hub.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var session hub.CLISession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", code).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCLINotFound
			}
			return err
		}

		if session.ExpiresAt.Before(time.Now().UTC()) && session.Status != "used" {
			if err := tx.Model(&session).Update("status", "expired").Error; err != nil {
				return err
			}
			return ErrCLICodeExpired
		}
		if session.Status == "pending" {
			return ErrCLIPending
		}
		if session.Status != "approved" || session.UserID == nil {
			return ErrCLINotApprovable
		}

		if err := tx.First(&outUser, *session.UserID).Error; err != nil {
			return err
		}
		if err := tx.Model(&session).Update("status", "used").Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &outUser, nil
}

func (s *AuthService) ResolveOrCreateUser(g *dto.GoogleIdentity, username string) (*hub.User, error) {
	if g == nil || strings.TrimSpace(g.Sub) == "" {
		return nil, errors.New("invalid google identity")
	}
	var identity hub.OAuthIdentity
	err := s.db.Where("provider = ? AND provider_user_id = ?", "google", g.Sub).First(&identity).Error
	if err == nil {
		// Google account is already linked to a ModelHub account
		// Don't allow re-signup; user should use signin instead
		return nil, ErrAccountAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return nil, ErrUsernameRequired
	}
	if !isValidUsername(username) {
		return nil, ErrInvalidUsername
	}

	var created hub.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		created = hub.User{
			Username:    username,
			Email:       strings.ToLower(strings.TrimSpace(g.Email)),
			DisplayName: g.Name,
		}
		if err := tx.Create(&created).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrUsernameAlreadyTaken
			}
			return err
		}
		oid := hub.OAuthIdentity{
			UserID:         created.ID,
			Provider:       "google",
			ProviderUserID: g.Sub,
		}
		if err := tx.Create(&oid).Error; err != nil {
			if isUniqueViolation(err) {
				return errOAuthIdentityExists
			}
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errOAuthIdentityExists) {
			// Another concurrent request already linked this Google identity.
			// Treat this as idempotent login and load the existing user.
			var existingIdentity hub.OAuthIdentity
			if err := s.db.Where("provider = ? AND provider_user_id = ?", "google", g.Sub).First(&existingIdentity).Error; err != nil {
				return nil, err
			}
			var existingUser hub.User
			if err := s.db.First(&existingUser, existingIdentity.UserID).Error; err != nil {
				return nil, err
			}
			return &existingUser, nil
		}
		return nil, err
	}
	return &created, nil
}

// ResolveExistingUser returns the user if their Google account is linked to a ModelHub account.
// Returns ErrAccountNotFound if the account doesn't exist.
func (s *AuthService) ResolveExistingUser(g *dto.GoogleIdentity) (*hub.User, error) {
	if g == nil || strings.TrimSpace(g.Sub) == "" {
		return nil, errors.New("invalid google identity")
	}
	var identity hub.OAuthIdentity
	err := s.db.Where("provider = ? AND provider_user_id = ?", "google", g.Sub).First(&identity).Error
	if err == nil {
		var user hub.User
		if err := s.db.First(&user, identity.UserID).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// Google account not linked to any ModelHub account
	return nil, ErrAccountNotFound
}

func isValidUsername(username string) bool {
	if len(username) < 1 || len(username) > 39 {
		return false
	}
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
