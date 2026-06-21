package utils

import (
	"bookstore-backend/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims are stored inside the access token.
//
// We include:
// - UserID: to identify the authenticated user.
// - Role: for authorization checks (admin/user).
type AccessClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

// RefreshClaims are stored inside the refresh token.
//
// We only need:
// - UserID: to know who owns the token.
type RefreshClaims struct {
	UserID uint `json:"user_id"`

	jwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived JWT.
//
// Used on:
// - Login
// - Refresh Token
func GenerateAccessToken(
	userID uint,
	role string,
) (string, error) {

	duration, err := time.ParseDuration(
		config.AppConfig.AccessTokenDuration,
	)
	if err != nil {
		return "", err
	}

	claims := AccessClaims{
		UserID: userID,
		Role:   role,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(duration),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(
		[]byte(config.AppConfig.JWTSecret),
	)
}

// GenerateRefreshToken creates a long-lived JWT.
//
// Used only to obtain a new access token.
func GenerateRefreshToken(
	userID uint,
) (string, time.Time, error) {

	duration, err := time.ParseDuration(
		config.AppConfig.RefreshTokenDuration,
	)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(duration)

	claims := RefreshClaims{
		UserID: userID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				expiresAt,
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte(config.AppConfig.JWTSecret),
	)

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func ValidateRefreshToken(
	tokenString string,
) (*RefreshClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(
				config.AppConfig.JWTSecret,
			), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)

	if !ok || !token.Valid {
		return nil,
			errors.New("invalid refresh token")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
// and returns its claims.
func ValidateAccessToken(
	tokenString string,
) (*AccessClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessClaims{},
		func(token *jwt.Token) (interface{}, error) {

			return []byte(
				config.AppConfig.JWTSecret,
			), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok :=
		token.Claims.(*AccessClaims)

	if !ok || !token.Valid {

		return nil,
			errors.New("invalid access token")
	}

	return claims, nil
}