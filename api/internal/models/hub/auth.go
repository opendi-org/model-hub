package hub

import (
	"time"

	"gorm.io/gorm"
)

// ── Users ─────────────────────────────────────────────────────────────────────
//
// User is the local account used by repository endpoints.
//
// MVP behavior:
// - users are created from Google callback + username
// - username is the public owner handle in route paths
// - username format validation happens in handlers/middleware
//
// Reserved fields:
// - DisplayName/Bio/AvatarURL are kept in the schema for future profile features.
//   MVP does not implement profile editing yet; they may be empty.

type User struct {
	gorm.Model
	Username    string `gorm:"type:text;uniqueIndex;not null"`
	Email       string `gorm:"type:text;uniqueIndex;not null"`
	DisplayName string `gorm:"type:text"`
	Bio         string `gorm:"type:text"`
	AvatarURL   string `gorm:"type:text"`
}

func (User) TableName() string { return "hub_users" }

// ── OAuth identities ──────────────────────────────────────────────────────────
//
// OAuthIdentity links an external identity to a local User.
// MVP only uses Provider="google" (one provider method).
// Provider tokens are not persisted.
//
// Notes:
// - UserID is unique in MVP, so one local user maps to one OAuth identity record.
// - ProviderUserID is the provider's stable subject identifier (e.g. OIDC "sub").
// - We intentionally do not store provider email here in MVP.

type OAuthIdentity struct {
	gorm.Model
	UserID         uint   `gorm:"uniqueIndex;not null;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Provider       string `gorm:"type:text;uniqueIndex:idx_oauth_identity;not null"` // e.g. "google"
	ProviderUserID string `gorm:"type:text;uniqueIndex:idx_oauth_identity;not null"` // stable sub / id from provider

	// Preload only
	User User `gorm:"foreignKey:UserID"`
}

func (OAuthIdentity) TableName() string { return "hub_oauth_identities" }

// Refresh-token persistence is intentionally not part of the current MVP.
// When refresh flow is implemented, add a dedicated model/table back here.

// ── CLI pre-auth approval sessions ────────────────────────────────────────────
//
// CLISession stores short-lived device login state for:
// - POST /v0/auth/cli/login (create pending session + code)
// - GET /v0/auth/login/google/start?cli_code=... + callback (approve)
// - POST /v0/auth/cli/poll (exchange approved session for access token)
//
// This table IS used in MVP (CLI flow is implemented).

type CLISession struct {
	gorm.Model
	Code      string    `gorm:"type:text;uniqueIndex;not null"`
	Status    string    `gorm:"type:text;default:'pending';not null;check:status IN ('pending','approved','used','expired')"` // pending|approved|used|expired
	UserID    *uint     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`                                               // Set on approval so poll can mint token for the correct user
	ExpiresAt time.Time `gorm:"type:timestamptz;not null"`

	// Preload only
	User *User `gorm:"foreignKey:UserID"`
}

func (CLISession) TableName() string { return "hub_cli_sessions" }
