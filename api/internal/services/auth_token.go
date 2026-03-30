package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"opendi.org/model-hub/api/internal/dto"
)

const googleOIDCIssuer = "https://accounts.google.com"

func IssueAccessToken(userID uint, username, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("missing jwt secret")
	}
	if ttl <= 0 {
		return "", errors.New("invalid token ttl")
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"uid":      userID,
		"username": username,
		"exp":      now.Add(ttl).Unix(),
		"iat":      now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyAccessToken(token, secret string) (*dto.AuthClaims, error) {
	if secret == "" {
		return nil, errors.New("missing jwt secret")
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	uidAny, ok := claims["uid"]
	if !ok {
		return nil, errors.New("invalid user id claim")
	}
	var uid uint64
	switch v := uidAny.(type) {
	case float64:
		if v <= 0 || math.Trunc(v) != v {
			return nil, errors.New("invalid user id claim")
		}
		uid = uint64(v)
	case int64:
		if v <= 0 {
			return nil, errors.New("invalid user id claim")
		}
		uid = uint64(v)
	case int:
		if v <= 0 {
			return nil, errors.New("invalid user id claim")
		}
		uid = uint64(v)
	default:
		return nil, errors.New("invalid user id claim")
	}
	if uid == 0 {
		return nil, errors.New("invalid user id claim")
	}
	out := &dto.AuthClaims{
		UserID: uint(uid),
	}
	if usernameAny, ok := claims["username"]; ok {
		if username, ok := usernameAny.(string); ok {
			out.Username = username
		}
	}
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		out.Exp = exp.Unix()
	}
	if iat, err := claims.GetIssuedAt(); err == nil && iat != nil {
		out.Iat = iat.Unix()
	}
	return out, nil
}

func BuildOAuthState(secret, mode, cliCode, username string) (string, error) {
	if secret == "" {
		return "", errors.New("missing oauth state secret")
	}
	if mode != "web" && mode != "cli" {
		return "", errors.New("invalid oauth mode")
	}
	nonce, err := randomHex(24)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"mode":    mode,
		"cliCode": cliCode,
		"username": username,
		"nonce":   nonce,
		"exp":     now.Add(10 * time.Minute).Unix(),
		"iat":     now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseOAuthState(state, secret string) (*dto.OAuthStateClaims, error) {
	if state == "" {
		return nil, errors.New("missing oauth state")
	}
	if secret == "" {
		return nil, errors.New("missing oauth state secret")
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(state, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	modeAny, ok := claims["mode"]
	if !ok {
		return nil, errors.New("invalid oauth state mode")
	}
	mode, ok := modeAny.(string)
	if !ok || (mode != "web" && mode != "cli") {
		return nil, errors.New("invalid oauth state mode")
	}
	nonceAny, ok := claims["nonce"]
	if !ok {
		return nil, errors.New("invalid oauth state nonce")
	}
	nonce, ok := nonceAny.(string)
	if !ok || nonce == "" {
		return nil, errors.New("invalid oauth state nonce")
	}
	var cliCode string
	if cliCodeAny, ok := claims["cliCode"]; ok {
		if v, ok := cliCodeAny.(string); ok {
			cliCode = v
		}
	}
	var username string
	if usernameAny, ok := claims["username"]; ok {
		if v, ok := usernameAny.(string); ok {
			username = v
		}
	}
	if mode == "cli" && cliCode == "" {
		return nil, errors.New("missing cli code in oauth state")
	}
	out := &dto.OAuthStateClaims{
		Mode:     mode,
		CLICode:  cliCode,
		Username: username,
		Nonce:    nonce,
	}
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		out.Exp = exp.Unix()
	}
	if iat, err := claims.GetIssuedAt(); err == nil && iat != nil {
		out.Iat = iat.Unix()
	}
	return out, nil
}

func BuildGoogleAuthorizeURL(clientID, redirectURL, state string) string {
	conf := googleOAuthConfig(clientID, "", redirectURL)
	return conf.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	)
}

func ExchangeGoogleCode(ctx context.Context, code, clientID, clientSecret, redirectURI string) (*dto.GoogleIdentity, error) {
	conf := googleOAuthConfig(clientID, clientSecret, redirectURI)
	client := &http.Client{Timeout: 10 * time.Second}
	ctxWithClient := context.WithValue(ctx, oauth2.HTTPClient, client)
	token, err := conf.Exchange(ctxWithClient, code)
	if err != nil {
		return nil, err
	}
	idToken, _ := token.Extra("id_token").(string)
	if idToken == "" {
		return nil, errors.New("google response missing id_token")
	}
	return verifyGoogleIDToken(ctx, idToken, clientID)
}

func verifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*dto.GoogleIdentity, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 10 * time.Second}
	ctx = oidc.ClientContext(ctx, client)

	provider, err := oidc.NewProvider(ctx, googleOIDCIssuer)
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	verifiedToken, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := verifiedToken.Claims(&claims); err != nil {
		return nil, err
	}
	if claims.Sub == "" || claims.Email == "" {
		return nil, errors.New("google id_token missing required identity claims")
	}
	if !claims.EmailVerified {
		return nil, errors.New("google account email is not verified")
	}
	return &dto.GoogleIdentity{
		Sub:   claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

func randomHex(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("invalid random hex length")
	}
	byteLen := n/2 + n%2
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(buf)
	if len(hexStr) > n {
		return hexStr[:n], nil
	}
	return hexStr, nil
}

func googleOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
