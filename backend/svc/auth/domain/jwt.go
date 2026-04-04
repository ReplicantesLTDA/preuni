// Package domain contains the core business logic for auth-svc.
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
)

// JWTService issues and manages JWT tokens.
type JWTService struct {
	signingKey            []byte
	accessTokenExpiry     time.Duration
	refreshTokenExpiryDays int
}

// NewJWTService creates a JWTService with the given signing key and expiry settings.
func NewJWTService(signingKey string, accessExpirySeconds int, refreshExpiryDays int) *JWTService {
	return &JWTService{
		signingKey:            []byte(signingKey),
		accessTokenExpiry:     time.Duration(accessExpirySeconds) * time.Second,
		refreshTokenExpiryDays: refreshExpiryDays,
	}
}

// IssueAccessToken creates a signed JWT access token for the given user.
func (s *JWTService) IssueAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := pkgmw.Claims{
		UserID:    userID,
		UserEmail: email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiry)),
			Issuer:    "preuni-auth",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", apperrors.Internal(err)
	}
	return signed, nil
}

// IssueRefreshToken generates a cryptographically random opaque token and its
// SHA-256 hash. Store the hash; return the raw token to the client.
func (s *JWTService) IssueRefreshToken() (rawToken, tokenHash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, apperrors.Internal(err)
	}
	rawToken = hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])
	expiresAt = time.Now().AddDate(0, 0, s.refreshTokenExpiryDays)
	return rawToken, tokenHash, expiresAt, nil
}

// ExpiresIn returns the access token expiry in seconds (for the API response).
func (s *JWTService) ExpiresIn() int {
	return int(s.accessTokenExpiry.Seconds())
}
