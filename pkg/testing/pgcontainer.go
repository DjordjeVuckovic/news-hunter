package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PGContainer struct {
	Container  testcontainers.Container
	ConnString string
}

type PGConfig struct {
	Database string
	Username string
	Password string
}

func NewPGContainer(ctx context.Context, cfg PGConfig) (*PGContainer, error) {
	return createPGContainer(ctx, cfg)
}

func NewPGContainerWithCleanup(ctx context.Context, tb testing.TB) *PGContainer {
	tb.Helper()

	container, err := createPGContainer(ctx, PGConfig{
		Database: "news_test_db",
		Username: "test",
		Password: "test",
	})
	if err != nil {
		tb.Fatalf("failed to create postgres container: %v", err)
	}

	tb.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container.Container); err != nil {
			tb.Logf("failed to terminate postgres container: %v", err)
		}
	})

	return container
}

func createPGContainer(ctx context.Context, cfg PGConfig) (*PGContainer, error) {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../..")
	migrationsDir := filepath.Join(projectRoot, "db", "migrations")

	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to find migration files: %w", err)
	}
	sort.Strings(migrationFiles)

	var initScript strings.Builder
	for i, f := range migrationFiles {
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", f, err)
		}
		initScript.Write(content)
		initScript.WriteString(";\n")
		if i < len(migrationFiles)-1 {
			initScript.WriteString("\n")
		}
	}

	tmpFile, err := os.CreateTemp("", "migrations-*.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.WriteString(initScript.String()); err != nil {
		return nil, fmt.Errorf("failed to write migrations: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:17.5",
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		postgres.WithInitScripts(tmpFile.Name()),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PGContainer{
		Container:  pgContainer,
		ConnString: connStr,
	}, nil
}
