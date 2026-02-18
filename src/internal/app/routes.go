package app

import (
	"etruscan/internal/api"
	"etruscan/internal/api/handlers"
	"etruscan/internal/domain/models"
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func (app *App) RegisterRoutes() {
	apiv1 := app.Echo.Group("/api/v1")

	apiv1.GET("/ready", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status": "ready",
		})
	})

	apiv1.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status": "healthy",
		})
	})

	authHdl := handlers.NewAuthHandler(app.Deps.AuthUseCase)

	apiv1.POST("/auth/login", authHdl.Login)

	protected := apiv1.Group("", echojwt.JWT([]byte(app.Config.JWTSecret)))
	admin := protected.Group("/admin", api.RequireRole(models.UserRoleAdmin))

	userHdl := handlers.NewUserHandler(app.Deps.UserUseCase)

	admin.GET("/users", userHdl.AdminList)
	admin.POST("/users", userHdl.AdminCreate)
	admin.GET("/users/:id", userHdl.AdminGetProfile)
	admin.PUT("/users/:id", userHdl.AdminUpdate)
	admin.DELETE("/users/:id", userHdl.AdminSoftDelete)

	approverGroupHdl := handlers.NewApproverGroupHandler(app.Deps.ApproverGroupUseCase)

	admin.POST("/approverGroups", approverGroupHdl.Create)
	admin.GET("/approverGroups/:id", approverGroupHdl.GetByID)
	admin.POST("/approverGroups/:id/members/add", approverGroupHdl.AddMembers)
	admin.POST("/approverGroups/:id/members/remove", approverGroupHdl.RemoveMembers)

	flagHdl := handlers.NewFlagHandler(app.Deps.FlagUseCase)

	protected.POST("/flags", flagHdl.Create)
	protected.GET("/flags", flagHdl.List)
	protected.GET("/flags/:id", flagHdl.GetByID)
	protected.PUT("/flags/:id", flagHdl.Update)
	protected.DELETE("/flags/:id", flagHdl.Delete)

	expHdl := handlers.NewExperimentHandler(app.Deps.ExperimentUseCase)

	protected.POST("/experiments", expHdl.Create)
	protected.GET("/experiments/:id", expHdl.GetByID)
}
