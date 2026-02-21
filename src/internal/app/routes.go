package app

import (
	"encoding/json"
	"etruscan/internal/api"
	"etruscan/internal/api/handlers"
	"etruscan/internal/domain/models"
	"io/fs"
	"net/http"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

func (app *App) RegisterRoutes() {
	specRoot, _ := fs.Sub(specFS, "spec")
	apiv1 := app.Echo.Group("/api/v1")
	apiv1.GET("/openapi.yaml", func(c echo.Context) error {
		data, err := fs.ReadFile(specRoot, "openapi.yaml")
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/x-yaml", data)
	})
	// /docs — Scalar UI (loads JS from CDN; use when online).
	app.Echo.GET("/docs", func(c echo.Context) error {
		data, err := fs.ReadFile(specRoot, "openapi.yaml")
		if err != nil {
			return err
		}
		var specMap map[string]interface{}
		if err := yaml.Unmarshal(data, &specMap); err != nil {
			return err
		}
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return err
		}
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecContent: string(specJSON),
			CustomOptions: scalar.CustomOptions{
				PageTitle: "ETRUSCAN A/B Platform API",
			},
			DarkMode: true,
		})
		if err != nil {
			return err
		}
		return c.HTMLBlob(http.StatusOK, []byte(htmlContent))
	})
	// /docs/offline — self-contained viewer (no CDN); use when no internet.
	app.Echo.GET("/docs/offline", func(c echo.Context) error {
		htmlContent, err := offlineDocsHTML(specRoot)
		if err != nil {
			return err
		}
		return c.HTMLBlob(http.StatusOK, htmlContent)
	})

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
	protected.GET("/experiments", expHdl.List)

	protected.GET("/experiments/:id", expHdl.GetByID)
	protected.PUT("/experiments/:id", expHdl.Update)
	protected.POST("/experiments/:id/sendOnReview", expHdl.SendOnReview)
	protected.POST("/experiments/:id/approve", expHdl.Approve)
	protected.POST("/experiments/:id/requestChanges", expHdl.RequestChanges)
	protected.POST("/experiments/:id/decline", expHdl.Decline)

	protected.GET("/experiments/:id/statusChanges", expHdl.ListStatusChanges)
	protected.GET("/experiments/:id/snapshots", expHdl.ListSnapshots)

	protected.POST("/experiments/:id/launch", expHdl.Launch)
	protected.POST("/experiments/:id/pause", expHdl.Pause)
	protected.POST("/experiments/:id/finish", expHdl.Finish)

	decideHdl := handlers.NewDecideHandler(app.Deps.DecideUseCase)
	eventsHdl := handlers.NewEventsHandler(app.Deps.EventsUseCase)

	apiv1.POST("/decide", decideHdl.Decide)
	apiv1.POST("/events", eventsHdl.Ingest)

	metricHdl := handlers.NewMetricHandler(app.Deps.MetricUseCase)
	protected.GET("/metrics", metricHdl.List)
	protected.POST("/metrics", metricHdl.Create)
	protected.GET("/metrics/:id", metricHdl.GetByID)

	reportHdl := handlers.NewReportHandler(app.Deps.ReportUseCase)
	protected.GET("/experiments/:id/report", reportHdl.GetExperimentReport)
}
