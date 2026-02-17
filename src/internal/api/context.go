package api

import (
	"errors"
	"etruscan/internal/domain/models"
	"etruscan/internal/provider"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var NoUserIdInContext = errors.New("no user id in context")

func ExtractUserAuthDataFromContext(c echo.Context) (models.UserAuthData, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return models.UserAuthData{}, NoUserIdInContext
	}
	return provider.ExtractUserAuthDataFromJWT(*token)
}
