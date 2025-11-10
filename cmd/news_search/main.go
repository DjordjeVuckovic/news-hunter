// Package main News Hunter API
// @title News Hunter API
// @version 1.0
// @description A full-text search engine for exploring multilingual news headlines and articles
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@newshunter.com
// @license.name Apache 2.0
// @license.url https://opensource.org/licenses/Apache-2.0
// @BasePath /
package main

import (
	"log/slog"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/router"
	"github.com/DjordjeVuckovic/news-hunter/internal/server"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	pkgserver "github.com/DjordjeVuckovic/news-hunter/pkg/server"
	"github.com/labstack/echo/v4"
)

func main() {
	sCfg, err := server.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	heathChecker := pkgserver.NewOkHealthChecker()

	s := server.New(sCfg, heathChecker).
		SetupMiddlewares().
		SetupHealthChecks().
		SetupOpenApi()

	s.Echo.GET("/", func(c echo.Context) error {
		return c.String(200, "News Hunter API is running")
	})

	appSettings := NewAppConfig()
	cfg, err := appSettings.Load()
	if err != nil {
		slog.Error("Failed to load app configuration", "error", err)
		os.Exit(1)
		return
	}

	reader, err := factory.NewSearcher(s.Context(), cfg.StorageConfig)
	if err != nil {
		slog.Error("Failed to create storage reader", "error", err)
		os.Exit(1)
		return
	}

	searchrouter := router.NewSearchRouter(s.Echo, reader)
	searchrouter.Bind()

	go func() {
		<-s.ShutdownSignal()
		slog.Info("Shutdown started, cleaning up resources...")
	}()

	err = s.Start()
	if err != nil {
		s.Echo.Logger.Error("Failed to start server: ", err)
		os.Exit(1)
	}
}
