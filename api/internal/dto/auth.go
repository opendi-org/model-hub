package dto

import "time"

type CLIPollRequest struct {
	Code string `json:"code" binding:"required"`
}

type CLILoginResponse struct {
	Code      string `json:"code"`
	LoginURL  string `json:"loginUrl"`
	ExpiresIn int64  `json:"expiresIn"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type MeResponse struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Internal auth token claims stored in access JWT.
type AuthClaims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username,omitempty"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// Internal OAuth state claims used during Google redirect/callback flow.
type OAuthStateClaims struct {
	Mode    string `json:"mode"`
	CLICode string `json:"cliCode,omitempty"`
	Username string `json:"username,omitempty"`
	Nonce   string `json:"nonce"`
	Exp     int64  `json:"exp"`
	Iat     int64  `json:"iat"`
}

// Google identity extracted from verified Google token response.
type GoogleIdentity struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
