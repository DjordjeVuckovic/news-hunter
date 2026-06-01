// Package version exposes the bench tool identity used in artifact provenance.
package version

import (
	"fmt"
	"sync"
)

// Semver is the semantic version. Bump on every release that changes artifact schemas or CLI.
const Semver = "1.0.0"

// Tool returns "bench/1.0.0" — the canonical identifier embedded in every
// artifact's meta block. SemVer alone is the reproducibility handle; callers
// who need the exact git revision should tag a release so the version is
// unambiguous without a SHA.
func Tool() string {
	toolOnce.Do(func() { toolStr = "bench/" + Semver })
	return toolStr
}

var (
	toolOnce sync.Once
	toolStr  string
)

// SchemaVersion is the current on-disk schema version for all bench artifacts.
// Bumped whenever any of (spec.yaml, suite.yaml, pool.yaml, annotations.yaml,
// report.json) changes shape. Loaders reject anything else.
const SchemaVersion = 1

// CheckSchema returns an error if got != SchemaVersion. Use in every loader.
func CheckSchema(got int, artifact string) error {
	if got == SchemaVersion {
		return nil
	}
	if got == 0 {
		return fmt.Errorf("%s: missing schema_version (expected %d) — regenerate with bench %s",
			artifact, SchemaVersion, regenHint(artifact))
	}
	return fmt.Errorf("%s: schema_version %d not supported (expected %d) — regenerate with bench %s",
		artifact, got, SchemaVersion, regenHint(artifact))
}

func regenHint(artifact string) string {
	switch artifact {
	case "spec", "suite":
		return "init"
	case "pool":
		return "pool"
	case "annotations":
		return "judge"
	case "report":
		return "run"
	default:
		return "init"
	}
}
