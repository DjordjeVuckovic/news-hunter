// Package trackctx resolves the "track" — the self-contained folder that
// owns a benchmark's spec, suite, pool, annotations, and reports. Every
// subcommand goes through Resolve (single track) or ResolveGlob (a group) to
// figure out which paths it operates on.
//
// Resolution precedence (highest wins):
//  1. Explicit --spec/--suite/--pool/--judgments/--output flags.
//  2. --track <name|path> flag.
//  3. Positional <track> argument.
//  4. Walk-up from cwd: search parent dirs for a track-shaped folder
//     (one that contains spec.yaml + suite.yaml + trec/).
//
// A track arg is one of:
//   - a verbatim filesystem path — absolute, ./- or ../-prefixed, or a *.yaml
//     etc. — used as-is (the escape hatch for tracks outside ./tracks).
//   - a track name → mapped under ./tracks/. Two layouts are supported:
//     flat (fts_quality → tracks/fts_quality) and nested with "/"
//     (news/fts → tracks/news/fts).
//   - a glob (news/*) → expanded by ResolveGlob across track-shaped matches.
//
// Grouping is explicit: only a glob fans out. A bare name always means exactly
// one track — there is no implicit "directory becomes a group" behaviour.
package trackctx

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	name       string // tracks-relative id, e.g. "fts_quality" or "news/fts"
	trecDir    string
	reportsDir string
}

// Resolve resolves a single track and layers any explicit overrides on top.
func Resolve(in Inputs) (*Track, error) {
	root, name, err := resolveRoot(in.TrackArg)
	if err != nil {
		return nil, err
	}
	t := newTrack(root, name)
	t.Spec = firstNonEmpty(in.SpecPath, t.Spec)
	t.Suite = firstNonEmpty(in.SuitePath, t.Suite)
	t.Pool = firstNonEmpty(in.PoolPath, t.Pool)
	return t, nil
}

// ResolveGlob expands a glob pattern (news/*) under tracks/ to every
// track-shaped match, sorted, naming each by its path relative to tracks/.
func ResolveGlob(pattern string) ([]*Track, error) {
	matches, err := filepath.Glob(filepath.Join(tracksDir, nameToRel(pattern)))
	if err != nil {
		return nil, fmt.Errorf("bad track pattern %q: %w", pattern, err)
	}
	sort.Strings(matches)
	tracks := make([]*Track, 0, len(matches))
	for _, m := range matches {
		abs, err := filepath.Abs(m)
		if err != nil || !isTrackShaped(abs) {
			continue
		}
		tracks = append(tracks, newTrack(canonicalise(abs), relName(m)))
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks match pattern %q under %s/", pattern, tracksDir)
	}
	return tracks, nil
}

// IsPattern reports whether arg is a glob (news/*) rather than a single name.
func IsPattern(arg string) bool {
	return strings.ContainsAny(arg, "*?[")
}

// newTrack builds a Track with all paths derived from the convention.
func newTrack(root, name string) *Track {
	t := &Track{
		Root:       root,
		name:       name,
		trecDir:    filepath.Join(root, trecDir),
		reportsDir: filepath.Join(root, reportsDir),
	}
	t.Spec = filepath.Join(root, specFile)
	t.Suite = filepath.Join(root, suiteFile)
	t.Pool = filepath.Join(t.trecDir, poolFile)
	return t
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

// Name returns the track's tracks-relative id (e.g. "fts_quality" or
// "news/fts") — used for display and to derive default spec.id.
func (t *Track) Name() string {
	if t.name != "" {
		return t.name
	}
	return filepath.Base(t.Root)
}

// resolveRoot does the precedence walk for the track folder itself, returning
// the absolute root and the tracks-relative name.
func resolveRoot(arg string) (root, name string, err error) {
	if arg != "" {
		return absTrackPath(arg)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("getwd: %w", err)
	}
	if r := walkUp(cwd); r != "" {
		r = canonicalise(r)
		return r, filepath.Base(r), nil
	}
	return "", "", fmt.Errorf(
		"no track specified and current directory is not inside a track folder.\n" +
			"  Pass a track name: bench <cmd> <track>\n" +
			"  Or scaffold one:   bench init <name>",
	)
}

func absTrackPath(arg string) (root, name string, err error) {
	if isVerbatimPath(arg) {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return "", "", fmt.Errorf("abs %q: %w", arg, err)
		}
		if !isTrackShaped(abs) {
			return "", "", fmt.Errorf("path %q is not a track folder (missing spec.yaml / suite.yaml / trec/)", arg)
		}
		abs = canonicalise(abs)
		return abs, filepath.Base(abs), nil
	}
	// Track name, flat or nested (news/fts). Map under tracks/.
	rel := nameToRel(arg)
	abs, err := filepath.Abs(filepath.Join(tracksDir, rel))
	if err != nil {
		return "", "", fmt.Errorf("abs %q: %w", rel, err)
	}
	if !isTrackShaped(abs) {
		return "", "", fmt.Errorf("track %q not found at %s (run: bench init %s)", arg, abs, arg)
	}
	return canonicalise(abs), rel, nil
}

// nameToRel maps a track name to its path relative to tracks/, tolerating a
// leading "tracks/" so that `fts_quality` and `tracks/fts_quality` (the form
// used by --track in the docs) resolve to the same track.
func nameToRel(arg string) string {
	rel := filepath.ToSlash(filepath.Clean(arg))
	return strings.TrimPrefix(rel, tracksDir+"/")
}

// relName strips the leading tracks/ segment from a matched path, yielding the
// tracks-relative track id (tracks/news/fts → news/fts).
func relName(match string) string {
	if rel, err := filepath.Rel(tracksDir, match); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.Base(match)
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

// isVerbatimPath reports whether a track arg is a filesystem path to use as-is,
// rather than a name mapped under tracks/. True for absolute paths, ./- or
// ../-prefixed paths, and anything carrying a recognised file extension. A
// plain or slash-nested name (news, news/fts) is NOT verbatim.
func isVerbatimPath(s string) bool {
	if s == "" {
		return false
	}
	if filepath.IsAbs(s) {
		return true
	}
	if s == "." || s == ".." ||
		strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") ||
		strings.HasPrefix(s, "."+string(filepath.Separator)) ||
		strings.HasPrefix(s, ".."+string(filepath.Separator)) {
		return true
	}
	return hasYAMLExt(s)
}

// isPath classifies a --judgments value: any separator or file extension means
// it's a path rather than a strategy name.
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
