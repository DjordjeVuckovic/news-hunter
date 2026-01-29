package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	openapi "github.com/DjordjeVuckovic/news-hunter/api/openapi-spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/apperr"
	mw "github.com/DjordjeVuckovic/news-hunter/internal/middleware"
	"github.com/DjordjeVuckovic/news-hunter/pkg/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const (
	DefaultGracefulShutdownTimeout = 10 * time.Second
)

type Server struct {
	*echo.Echo

	cfg *Config

	checker server.HealthChecker

	ctx context.Context

	gracefulShutdownTimeout time.Duration
	shutdownSig             chan struct{}
}

func New(cfg *Config, checker server.HealthChecker) *Server {
	e := echo.New()

	e.DisableHTTP2 = !cfg.UseHttp2

	s := &Server{
		Echo:                    e,
		cfg:                     cfg,
		checker:                 checker,
		ctx:                     context.Background(),
		gracefulShutdownTimeout: DefaultGracefulShutdownTimeout,
		shutdownSig:             make(chan struct{}),
	}

	return s
}

func (s *Server) Context() context.Context {
	return s.ctx
}

func (s *Server) ShutdownSignal() chan struct{} {
	return s.shutdownSig
}

func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := s.Echo.Start(":" + s.cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Echo.Logger.Fatal("shutting down the server")
		}
	}()

	s.ctx = ctx

	<-ctx.Done()

	close(s.shutdownSig)

	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTimeout)
	defer cancel()

	if err := s.Echo.Shutdown(ctx); err != nil {
		s.Echo.Logger.Fatal(err)
		return err
	}
	slog.Info("Server shut down gracefully ...")

	return nil
}

func (s *Server) SetupMiddlewares() *Server {
	s.Echo.Use(middleware.RequestID())
	s.Echo.Use(mw.Logger())
	s.Echo.Use(middleware.Recover())
	s.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: s.cfg.CorsOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	return s
}

func (s *Server) SetupHealthChecks(path string) *Server {
	s.Echo.GET(path, s.handleHealthCheck)

	return s
}

func (s *Server) handleHealthCheck(c echo.Context) error {
	if !s.checker.Healthy(c.Request().Context()) {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) SetupErrorHandler() *Server {
	s.Echo.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			_ = c.JSON(http.StatusBadRequest, map[string]string{"error": ve.Message})
			return
		}

		var he *echo.HTTPError
		if errors.As(err, &he) {
			msg := fmt.Sprintf("%v", he.Message)
			_ = c.JSON(he.Code, map[string]string{"error": msg})
			return
		}

		slog.Error("Unhandled error", "error", err)
		_ = c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return s
}

func (s *Server) SetupOpenApi(path string) *Server {
	openapi.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", s.cfg.Port)

	s.Echo.GET(path, echoSwagger.WrapHandler)

	return s
}
