package server

import (
	"errors"
	"fmt"
	"github.com/DjordjeVuckovic/news-hunter/pkg/stringsutil"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	UseHttp2    bool
	CorsOrigins []string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("cmd/news_search/.env")
	if err != nil {
		slog.Info("Skipping .env ...", "error", err)
	}

	useHttp2Str := os.Getenv("USE_HTTP2")
	useHttp2 := useHttp2Str == "true"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := validatePort(port); err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	var origins []string
	corsOriginsEnv := os.Getenv("CORS_ORIGINS")
	if corsOriginsEnv != "" {
		origins = strings.Split(corsOriginsEnv, ",")
		// Trim whitespace from each origin
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		// Remove empty origins
		origins = stringsutil.RemoveEmptyStrings(origins)
	}

	if len(origins) == 0 {
		origins = []string{"*"}
	}

	return &Config{
		Port:        port,
		UseHttp2:    useHttp2,
		CorsOrigins: origins,
	}, nil
}

func validatePort(port string) error {
	portNum, err := strconv.Atoi(port)

	if err != nil {
		return errors.New("port must be a number")
	}

	if portNum < 1 || portNum > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}
