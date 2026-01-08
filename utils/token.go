package utils

import (
	"livo-fiber-backend/config"
	"time"

	"aidanwoods.dev/go-paseto"
)

type TokenClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func GenerateAccessToken(claims TokenClaims, cfg *config.Config) (string, error) {
	location, _ := time.LoadLocation(cfg.DbTz)
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now().In(location))
	token.SetNotBefore(time.Now().In(location))
	token.SetExpiration(time.Now().In(location).Add(time.Duration(cfg.AccessTokenTTL) * time.Minute))
	token.SetString("userId", claims.UserID)
	token.SetString("username", claims.Username)
	token.Set("roles", claims.Roles)
	token.SetString("type", "access")

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(cfg.PasetoSymmetricKey))
	if err != nil {
		return "", err
	}
	return token.V4Encrypt(key, nil), nil
}

func GenerateRefreshToken(claims TokenClaims, cfg *config.Config) (string, error) {
	location, _ := time.LoadLocation(cfg.DbTz)
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now().In(location))
	token.SetNotBefore(time.Now().In(location))
	token.SetExpiration(time.Now().In(location).Add(time.Duration(cfg.RefreshTokenTTL) * 24 * time.Hour))
	token.SetString("userId", claims.UserID)
	token.SetString("username", claims.Username)
	token.SetString("type", "refresh")

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(cfg.PasetoSymmetricKey))
	if err != nil {
		return "", err
	}

	return token.V4Encrypt(key, nil), nil
}

func ValidateToken(tokenString string, cfg *config.Config) (*paseto.Token, error) {
	key, err := paseto.V4SymmetricKeyFromBytes([]byte(cfg.PasetoSymmetricKey))
	if err != nil {
		return nil, err
	}

	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(key, tokenString, nil)
	return token, err
}
