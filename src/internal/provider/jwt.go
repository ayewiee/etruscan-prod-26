package provider

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTProvider struct {
	jwtSecret []byte
	ttl       time.Duration
}

func NewJWTProvider(jwtSecret []byte, ttl time.Duration) *JWTProvider {
	return &JWTProvider{jwtSecret: jwtSecret, ttl: ttl}
}

func (j *JWTProvider) GetTTL() time.Duration {
	return j.ttl
}

func (j *JWTProvider) GenerateToken(data models.UserAuthData) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  data.ID.String(),
		"role": string(data.Role),
		"iat":  now.Unix(),
		"exp":  now.Add(j.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.jwtSecret)
}

func ExtractUserAuthDataFromJWT(token jwt.Token) (models.UserAuthData, error) {
	claims := token.Claims.(jwt.MapClaims)
	userId, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return models.UserAuthData{}, err
	}
	return models.UserAuthData{
		ID:   userId,
		Role: models.UserRole(claims["role"].(string)),
	}, nil
}
