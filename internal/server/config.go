package server

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	UseHttp2    bool
	CorsOrigins []string
}

func LoadConfig() (*Config, error) {
	useHttp2Str := os.Getenv("USE_HTTP2")
	useHttp2 := useHttp2Str == "true"

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	origins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	return &Config{
		Port:        port,
		UseHttp2:    useHttp2,
		CorsOrigins: origins,
	}, nil
}
