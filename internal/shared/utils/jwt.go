package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ─── Claims ────────────────────────────────────────────────────────────────────

// JWTClaims custom claims untuk JWT
type JWTClaims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	IsSuperadmin bool   `json:"is_superadmin"`
	IsStaff      bool   `json:"is_staff"`
	TokenType    string `json:"token_type"` // access | refresh
	jwt.RegisteredClaims
}

// ─── JWT Manager ───────────────────────────────────────────────────────────────

type JWTManager struct {
	secret        []byte
	issuer        string
	accessExpMin  int
	refreshExpDay int
}

func NewJWTManager(secret, issuer string, accessExpMin, refreshExpDay int) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		issuer:        issuer,
		accessExpMin:  accessExpMin,
		refreshExpDay: refreshExpDay,
	}
}

// GenerateAccessToken membuat access token JWT
func (j *JWTManager) GenerateAccessToken(userID int64, username string, isSuperadmin, isStaff bool) (string, error) {
	claims := JWTClaims{
		UserID:       userID,
		Username:     username,
		IsSuperadmin: isSuperadmin,
		IsStaff:      isStaff,
		TokenType:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.accessExpMin) * time.Minute)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken membuat refresh token JWT dengan JTI unik
func (j *JWTManager) GenerateRefreshToken(userID int64, username string) (string, string, time.Time, error) {
	jti := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(j.refreshExpDay) * 24 * time.Hour)

	claims := JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return tokenStr, jti, expiresAt, nil
}

// ParseToken mem-parse dan memvalidasi JWT token
func (j *JWTManager) ParseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	}, jwt.WithIssuer(j.issuer))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AccessExpSeconds mengembalikan durasi access token dalam detik
func (j *JWTManager) AccessExpSeconds() int {
	return j.accessExpMin * 60
}

// RefreshExpDuration mengembalikan durasi refresh token
func (j *JWTManager) RefreshExpDuration() time.Duration {
	return time.Duration(j.refreshExpDay) * 24 * time.Hour
}
