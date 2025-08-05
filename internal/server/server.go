package server

import (
	"context"
	"errors"
	mw "github.com/DjordjeVuckovic/news-hunter/internal/middleware"
	"github.com/DjordjeVuckovic/news-hunter/pkg/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	GracefulShutdownTimeout = 10 * time.Second
)

type Server struct {
	*echo.Echo

	cfg *Config

	checker server.HealthChecker
}

func New(cfg *Config, checker server.HealthChecker) *Server {
	e := echo.New()

	e.DisableHTTP2 = !cfg.UseHttp2

	s := &Server{
		Echo:    e,
		cfg:     cfg,
		checker: checker,
	}

	return s
}

func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := s.Echo.Start(":" + s.cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Echo.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
	defer cancel()

	if err := s.Echo.Shutdown(ctx); err != nil {
		s.Echo.Logger.Fatal(err)
		return err
	}
	return nil
}

func (s *Server) SetupMiddlewares() *Server {
	s.Echo.Use(mw.Logger())
	s.Echo.Use(middleware.Recover())
	s.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: s.cfg.CorsOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	return s
}

func (s *Server) SetupHealthChecks() *Server {
	s.Echo.GET("/health", s.handleHealthCheck)

	return s
}

func (s *Server) handleHealthCheck(c echo.Context) error {
	if !s.checker.Healthy(c.Request().Context()) {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}
