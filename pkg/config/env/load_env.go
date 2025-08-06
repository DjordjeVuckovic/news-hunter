package env

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads environment variables from a .env file.
// It uses the ENV_PATH environment variable to determine the path to the .env file.
func LoadDotEnv(env string, defaultPath string) error {
	var envPath string
	if os.Getenv("ENV_PATH") != "" {
		envPath = os.Getenv("ENV_PATH")
	} else {
		slog.Info("ENV_PATH is not set, using default path", "defaultPath", defaultPath)
		envPath = defaultPath
	}

	err := godotenv.Load(envPath)
	if err != nil {
		if env == "local" || env == "" {
			slog.Error("Failed to load environment variables in local mode", "error", err)
			return err
		}
		slog.Debug("Skipping .env ...")
	}

	return nil
}
