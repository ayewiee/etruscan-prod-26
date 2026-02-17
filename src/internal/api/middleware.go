package api

import (
	"etruscan/internal/domain/models"

	"github.com/labstack/echo/v4"
)

func RequireRole(allowedRoles ...models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authData, err := ExtractUserAuthDataFromContext(c)
			if err != nil {
				return models.ErrUnauthorized
			}

			for _, role := range allowedRoles {
				if authData.Role == role {
					return next(c)
				}
			}

			return models.ErrForbidden
		}
	}
}
