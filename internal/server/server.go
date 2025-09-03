package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	openapi "github.com/DjordjeVuckovic/news-hunter/api/openapi-spec"
	mw "github.com/DjordjeVuckovic/news-hunter/internal/middleware"
	"github.com/DjordjeVuckovic/news-hunter/pkg/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const (
	GracefulShutdownTimeout = 10 * time.Second
)

type Server struct {
	*echo.Echo

	cfg *Config

	checker server.HealthChecker

	Ctx context.Context
}

func New(cfg *Config, checker server.HealthChecker) *Server {
	e := echo.New()

	e.DisableHTTP2 = !cfg.UseHttp2

	s := &Server{
		Echo:    e,
		cfg:     cfg,
		checker: checker,
		Ctx:     context.Background(),
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

	s.Ctx = ctx

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

func (s *Server) SetupHealthChecker() *Server {
	s.Echo.GET("/health", s.handleHealthCheck)

	return s
}

func (s *Server) handleHealthCheck(c echo.Context) error {
	if !s.checker.Healthy(c.Request().Context()) {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) SetupOpenApi() *Server {
	openapi.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", s.cfg.Port)

	s.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	return s
}
