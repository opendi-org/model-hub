package hub

import (
	"time"

	"gorm.io/gorm"
)

// ── Users ─────────────────────────────────────────────────────────────────────
//
// A User is a fully-realised hub identity. No partial or pending state is
// stored here — a User row is only written after both the OAuth handshake and
// username selection are complete.
//
// The pending onboarding state lives in a short-lived signed JWT issued by the
// OAuth callback (claim: stage="pending_username"). The frontend holds this JWT
// and exchanges it — along with the chosen username — in a single request that
// atomically INSERTs the User row and the linked OAuthIdentity row.
//
// Username is the public handle used in :owner path segments and must be
// URL-safe ([a-zA-Z0-9_-], max 39 chars). Validated in application code.
//
// Bio and AvatarURL are reserved for future profile features. AvatarURL is
// always user-supplied — we never mirror the Google profile photo.

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
// One row per (Provider, ProviderUserID) pair, linked to a User.
// Currently only Provider="google" is used. Adding GitHub or OIDC later is a
// data change only — no schema migration needed.
//
// We do NOT store provider access or refresh tokens. They are only needed
// during the OAuth handshake to read the user's identity and are discarded
// immediately after. The hub issues its own tokens (see RefreshToken).
//
// Email records the provider-side address at the time of last login. It may
// diverge from User.Email if the user later updates their hub email.

type OAuthIdentity struct {
	gorm.Model
	UserID         uint   `gorm:"not null"`
	Provider       string `gorm:"type:text;uniqueIndex:idx_oauth_identity;not null"` // e.g. "google"
	ProviderUserID string `gorm:"type:text;uniqueIndex:idx_oauth_identity;not null"` // stable sub / id from provider
	Email          string `gorm:"type:text"`                                         // provider-side email at last login

	// Preload only
	User User `gorm:"foreignKey:UserID"`
}

func (OAuthIdentity) TableName() string { return "hub_oauth_identities" }

// ── Refresh tokens ────────────────────────────────────────────────────────────
//
// One row per active refresh token issued by the hub. Storing tokens
// server-side (rather than relying solely on JWT expiry) enables:
//
//   Revocation     SET revoked_at = NOW() → token invalid immediately.
//   Logout         DELETE WHERE user_id=? AND device_id=?
//   Revoke all     DELETE WHERE user_id=?
//   Key rotation   Rotate signing keys without invalidating live sessions.
//   Replay detect  After rotation, if the old hash appears again it was stolen
//                  → revoke all tokens in the same FamilyID chain.
//
// Token flow:
//   Login    → issue short-lived access JWT + opaque refresh token → INSERT row.
//   API call → verify access JWT signature + expiry locally (no DB hit).
//   Refresh  → SELECT WHERE token_hash=?, check not expired or revoked,
//              issue new access JWT + new refresh token (same FamilyID),
//              DELETE old row, INSERT new row.
//   Logout   → DELETE row for this device.
//
// TokenHash   SHA-256(raw_token_bytes), hex-encoded. Raw token bytes are never
//             persisted.
//
// FamilyID    Random UUID assigned at initial login, inherited by every rotated
//             successor. Replay detection revokes all rows WHERE family_id=?.
//
// DeviceID    Opaque client-supplied string (e.g. stable machine fingerprint for
//             CLI, browser fingerprint for web). Informational — used for active-
//             sessions display and single-device revocation, not cryptographic
//             binding.
//
// RevokedAt   NULL = active token. Set on revocation (not immediately deleted)
//             so replay detection has a window to detect re-use of a rotated
//             token. A background job hard-deletes revoked rows older than 24h.

type RefreshToken struct {
	gorm.Model
	UserID     uint       `gorm:"not null"`
	TokenHash  string     `gorm:"type:text;uniqueIndex;not null"`
	FamilyID   string     `gorm:"type:uuid;not null"`
	DeviceID   string     `gorm:"type:text"`
	DeviceName string     `gorm:"type:text"` // human label, e.g. "CLI – dev-macbook"
	ExpiresAt  time.Time  `gorm:"type:timestamptz;not null"`
	RevokedAt  *time.Time `gorm:"type:timestamptz"` // NULL = active

	// Preload only
	User User `gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string { return "hub_refresh_tokens" }

// ── CLI sessions (device-code flow) ──────────────────────────────────────────
//
// Supports the CLI device-code login: POST /auth/cli/login → POST /auth/cli/poll.
//
// Lifecycle:
//   POST /auth/cli/login  → INSERT row (Status="pending")
//                         → return {device_code, user_code, login_url, expires_in}
//   User opens browser    → confirms the UserCode displayed in the URL
//                         → UPDATE SET status="approved", user_id=<id>
//   POST /auth/cli/poll   → finds Status="approved"
//                         → issues access JWT + RefreshToken row for the device
//                         → UPDATE SET status="used"
//   Background GC         → DELETE WHERE status IN ('used','expired')
//                           AND created_at < NOW() - INTERVAL '10 minutes'
//
// DeviceCode  Cryptographically random secret polled by the CLI. Never shown in
//             the browser.
// UserCode    Short human-readable code (e.g. "ABCD-1234") embedded in the
//             browser confirmation URL so the user can verify the right device.

type CLISession struct {
	gorm.Model
	DeviceCode string    `gorm:"type:text;uniqueIndex;not null"`
	UserCode   string    `gorm:"type:text;uniqueIndex;not null"`
	Status     string    `gorm:"type:text;default:'pending';not null"` // pending|approved|used|expired
	UserID     *uint     // NULL until the user approves in the browser
	ExpiresAt  time.Time `gorm:"type:timestamptz;not null"`

	// Preload only
	User *User `gorm:"foreignKey:UserID"`
}

func (CLISession) TableName() string { return "hub_cli_sessions" }

