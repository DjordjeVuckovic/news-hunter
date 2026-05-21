// Package trackctx resolves the "track" — the self-contained folder that
// owns a benchmark's spec, suite, pool, annotations, and reports. Every
// subcommand goes through Resolve() to figure out which paths it operates on.
//
// Resolution precedence (highest wins):
//  1. Explicit --spec/--suite/--pool/--judgments/--output flags.
//  2. --track <name|path> flag.
//  3. Positional <track> argument.
//  4. Walk-up from cwd: search parent dirs for a track-shaped folder
//     (one that contains spec.yaml + suite.yaml + trec/).
//
// A "track name" without slashes maps to ./tracks/<name>/. A path-like value
// (containing /) is used verbatim.
package trackctx

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	tracksDir    = "tracks"
	specFile     = "spec.yaml"
	suiteFile    = "suite.yaml"
	trecDir      = "trec"
	poolFile     = "pool.yaml"
	reportsDir   = "reports"
	latestReport = "latest.json"
)

// Inputs lets callers pass explicit overrides that beat track inference.
// Any field left blank falls back to the track convention.
type Inputs struct {
	TrackArg   string // --track value OR positional arg (already merged by caller)
	SpecPath   string // --spec override
	SuitePath  string // --suite override
	PoolPath   string // --pool override
	OutputPath string // --output override
	Judgments  string // --judgments value: "lexical" (strategy name) | path/to.yaml | empty
}

// Track holds absolute paths derived from a track folder + overrides.
type Track struct {
	Root  string // absolute path to the track folder
	Spec  string
	Suite string
	Pool  string
	// JudgmentsForStrategy returns the conventional path for an annotations
	// file produced by the given strategy, inside this track.
	trecDir    string
	reportsDir string
}

// Resolve performs the precedence walk described in the package doc and
// returns a Track or a descriptive error suggesting the next step.
func Resolve(in Inputs) (*Track, error) {
	root, err := resolveRoot(in.TrackArg)
	if err != nil {
		return nil, err
	}
	t := &Track{
		Root:       root,
		trecDir:    filepath.Join(root, trecDir),
		reportsDir: filepath.Join(root, reportsDir),
	}
	t.Spec = firstNonEmpty(in.SpecPath, filepath.Join(root, specFile))
	t.Suite = firstNonEmpty(in.SuitePath, filepath.Join(root, suiteFile))
	t.Pool = firstNonEmpty(in.PoolPath, filepath.Join(t.trecDir, poolFile))
	return t, nil
}

// JudgmentsPath resolves --judgments. Three shapes:
//   - "" or absent → track's default (caller should consult spec.defaults).
//   - bare name like "lexical" → trec/annotations.lexical.yaml.
//   - path containing / or .yaml → used verbatim.
func (t *Track) JudgmentsPath(value string) string {
	if value == "" {
		return ""
	}
	if isPath(value) {
		return value
	}
	return filepath.Join(t.trecDir, "annotations."+value+".yaml")
}

// QrelsPath mirrors JudgmentsPath for TREC qrels exports.
func (t *Track) QrelsPath(value string) string {
	if value == "" {
		return ""
	}
	if isPath(value) {
		return value
	}
	return filepath.Join(t.trecDir, "qrels."+value+".tsv")
}

// ReportPath returns tracks/<name>/reports/<run_id>.json. Callers pass the
// run_id from meta.NewRunID("run").
func (t *Track) ReportPath(runID string) string {
	return filepath.Join(t.reportsDir, runID+".json")
}

// LatestReportPath is the conventional pointer to the most-recent report.
func (t *Track) LatestReportPath() string {
	return filepath.Join(t.reportsDir, latestReport)
}

// Name returns the track folder's basename — used to derive default spec.id.
func (t *Track) Name() string {
	return filepath.Base(t.Root)
}

// resolveRoot does the precedence walk for the track folder itself.
func resolveRoot(arg string) (string, error) {
	if arg != "" {
		return absTrackPath(arg)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	if root := walkUp(cwd); root != "" {
		return canonicalise(root), nil
	}
	return "", fmt.Errorf(
		"no track specified and current directory is not inside a track folder.\n" +
			"  Pass a track name: bench <cmd> <track>\n" +
			"  Or scaffold one:   bench init <name>",
	)
}

func absTrackPath(arg string) (string, error) {
	var raw string
	if isPath(arg) {
		raw = arg
	} else {
		raw = filepath.Join(tracksDir, arg)
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", raw, err)
	}
	if !isTrackShaped(abs) {
		if isPath(arg) {
			return "", fmt.Errorf("path %q is not a track folder (missing spec.yaml / suite.yaml / trec/)", arg)
		}
		return "", fmt.Errorf("track %q not found at %s (run: bench init %s)", arg, abs, arg)
	}
	return canonicalise(abs), nil
}

// canonicalise resolves symlinks so a track's Root is stable across OS quirks
// (notably macOS where /tmp -> /private/tmp). Failure falls back to the input.
func canonicalise(p string) string {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real
	}
	return p
}

// walkUp searches cwd and its parents for a track-shaped folder. Stops at the
// filesystem root or after 16 hops, whichever first.
func walkUp(start string) string {
	dir := start
	for i := 0; i < 16; i++ {
		if isTrackShaped(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

// isTrackShaped checks for the three signature artifacts of a track folder.
func isTrackShaped(dir string) bool {
	if !isDir(dir) {
		return false
	}
	return isFile(filepath.Join(dir, specFile)) &&
		isFile(filepath.Join(dir, suiteFile)) &&
		isDir(filepath.Join(dir, trecDir))
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func isPath(s string) bool {
	if s == "" {
		return false
	}
	return filepath.IsAbs(s) || containsSeparator(s) || hasYAMLExt(s)
}

func containsSeparator(s string) bool {
	for _, r := range s {
		if r == '/' || r == filepath.Separator {
			return true
		}
	}
	return false
}

func hasYAMLExt(s string) bool {
	ext := filepath.Ext(s)
	return ext == ".yaml" || ext == ".yml" || ext == ".tsv" || ext == ".json"
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
