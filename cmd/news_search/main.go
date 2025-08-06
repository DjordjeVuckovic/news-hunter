package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/router"
	"github.com/DjordjeVuckovic/news-hunter/internal/server"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
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
		SetupHealthChecks().
		SetupMiddlewares()

	s.Echo.GET("/", func(c echo.Context) error {
		return c.String(200, "News Hunter API is running")
	})

	appSettings := NewAppSettings()
	cfg, err := appSettings.Load()
	if err != nil {
		slog.Error("Failed to load app configuration", "error", err)
		os.Exit(1)
	}

	reader, err := newReader(s.Ctx, cfg)
	if err != nil {
		slog.Error("Failed to create storage reader", "error", err)
		os.Exit(1)
	}

	searchrouter := router.NewSearchRouter(s.Echo, reader)
	searchrouter.Bind()

	err = s.Start()
	if err != nil {
		s.Echo.Logger.Error("Failed to start server: ", err)
		os.Exit(1)
	}
}

func newReader(ctx context.Context, cfg *NewsSearchConfig) (storage.Reader, error) {
	storageType := cfg.StorageType
	var reader storage.Reader
	var err error

	switch cfg.StorageType {
	case storage.ES:
		reader, err = factory.NewReader(cfg.StorageType, ctx, *cfg.Elasticsearch)
	case storage.PG:
		reader, err = factory.NewReader(cfg.StorageType, ctx, *cfg.Postgres)
	}
	if err != nil {
		slog.Error("failed to create storer", "error", err)
		return nil, err
	}
	return reader, fmt.Errorf("unsupported storage type: %s", storageType)
}
