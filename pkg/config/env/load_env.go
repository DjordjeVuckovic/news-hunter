package env

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads environment variables from a .env file.
// It uses the ENV_PATH environment variable to determine the path to the .env file.
func LoadDotEnv(env string, paths ...string) error {
	var envPath string
	if os.Getenv("ENV_PATHS") != "" {
		envPath = os.Getenv("ENV_PATHS")
	} else {
		slog.Info("ENV_PATHS is not set, using only provided paths", "paths", paths)
	}

	decodedPaths := strings.Split(envPath, ",")

	err := godotenv.Load(append(decodedPaths, paths...)...)
	if err != nil {
		if env == "local" || env == "" {
			slog.Info("Failed to load environment variables in local mode", "error", err)
			return err
		}
		slog.Debug("Skipping .env ...")
	}

	return nil
}
