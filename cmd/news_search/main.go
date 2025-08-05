package main

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/server"
	pkgserver "github.com/DjordjeVuckovic/news-hunter/pkg/server"
	"github.com/labstack/echo/v4"
	"log/slog"
	"os"
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

	err = s.Start()
	if err != nil {
		s.Echo.Logger.Error("Failed to start server: ", err)
		os.Exit(1)
	}
}
