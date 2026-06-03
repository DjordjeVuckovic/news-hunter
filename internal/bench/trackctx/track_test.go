package trackctx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTrack writes the three signature files that mark a folder as a track.
func makeTrack(t *testing.T, root string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "trec"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "spec.yaml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "suite.yaml"), []byte(""), 0644))
}

// canonical returns the EvalSymlinks form of p — assertions about paths
// must compare against this since Resolve canonicalises its output.
func canonical(t *testing.T, p string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(p)
	require.NoError(t, err)
	return real
}

func TestResolve_BareNameUsesTracksDir(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "smoke")
	makeTrack(t, track)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	tr, err := Resolve(Inputs{TrackArg: "smoke"})
	require.NoError(t, err)
	assert.Equal(t, canonical(t, track), tr.Root)
	canonTrack := canonical(t, track)
	assert.Equal(t, filepath.Join(canonTrack, "spec.yaml"), tr.Spec)
	assert.Equal(t, filepath.Join(canonTrack, "suite.yaml"), tr.Suite)
	assert.Equal(t, filepath.Join(canonTrack, "trec", "pool.yaml"), tr.Pool)
}

func TestResolve_PathArgUsedVerbatim(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "weird", "place")
	makeTrack(t, track)

	tr, err := Resolve(Inputs{TrackArg: track})
	require.NoError(t, err)
	assert.Equal(t, canonical(t, track), tr.Root)
}

func TestResolve_WalkUpFromNestedCWD(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "demo")
	makeTrack(t, track)
	nested := filepath.Join(track, "trec")
	require.NoError(t, os.MkdirAll(nested, 0755))

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(nested))

	tr, err := Resolve(Inputs{})
	require.NoError(t, err)
	assert.Equal(t, canonical(t, track), tr.Root)
}

func TestResolve_ExplicitFlagsOverride(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "demo")
	makeTrack(t, track)

	tr, err := Resolve(Inputs{
		TrackArg: track,
		SpecPath: "/abs/elsewhere/spec.yaml",
		PoolPath: "/abs/elsewhere/pool.yaml",
	})
	require.NoError(t, err)
	assert.Equal(t, "/abs/elsewhere/spec.yaml", tr.Spec)
	assert.Equal(t, "/abs/elsewhere/pool.yaml", tr.Pool)
	// Suite NOT overridden falls back to the track convention.
	assert.Equal(t, filepath.Join(canonical(t, track), "suite.yaml"), tr.Suite)
}

func TestResolve_UnknownTrackErrors(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	_, err := Resolve(Inputs{TrackArg: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bench init nope", "error should suggest init command")
}

func TestResolve_NoTrackNoWalkUp(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	_, err := Resolve(Inputs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no track specified")
}

func TestResolve_TracksPrefixedNameIsBackwardCompatible(t *testing.T) {
	// The docs use `--track tracks/fts_quality`; that must resolve to the same
	// track as the bare name, not tracks/tracks/fts_quality.
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "fts_quality")
	makeTrack(t, track)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	tr, err := Resolve(Inputs{TrackArg: "tracks/fts_quality"})
	require.NoError(t, err)
	assert.Equal(t, canonical(t, track), tr.Root)
	assert.Equal(t, "fts_quality", tr.Name())
}

func TestResolve_NestedSingleName(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "news", "fts")
	makeTrack(t, track)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	tr, err := Resolve(Inputs{TrackArg: "news/fts"})
	require.NoError(t, err)
	assert.Equal(t, canonical(t, track), tr.Root)
	assert.Equal(t, "news/fts", tr.Name(), "nested name carries the dataset prefix")
}

func TestResolve_BareDatasetDirIsError(t *testing.T) {
	// "news" is a directory of tracks but not itself a track. Grouping is
	// explicit (via glob), so a bare name must NOT implicitly expand — it errors.
	dir := t.TempDir()
	makeTrack(t, filepath.Join(dir, "tracks", "news", "fts"))
	makeTrack(t, filepath.Join(dir, "tracks", "news", "fuzzy"))

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	_, err := Resolve(Inputs{TrackArg: "news"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveGlob_Expands(t *testing.T) {
	dir := t.TempDir()
	makeTrack(t, filepath.Join(dir, "tracks", "news", "fts"))
	makeTrack(t, filepath.Join(dir, "tracks", "news", "fuzzy"))
	makeTrack(t, filepath.Join(dir, "tracks", "wiki", "fts"))
	// A glob match that isn't track-shaped must be skipped.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "tracks", "news", "scratch"), 0755))

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	tracks, err := ResolveGlob("news/*")
	require.NoError(t, err)

	var names []string
	for _, tr := range tracks {
		names = append(names, tr.Name())
	}
	assert.Equal(t, []string{"news/fts", "news/fuzzy"}, names, "track-shaped matches only, sorted")
}

func TestResolveGlob_NoMatchErrors(t *testing.T) {
	dir := t.TempDir()
	makeTrack(t, filepath.Join(dir, "tracks", "news", "fts"))

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))

	_, err := ResolveGlob("wiki/*")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tracks match pattern")
}

func TestIsPattern(t *testing.T) {
	assert.True(t, IsPattern("news/*"))
	assert.True(t, IsPattern("news/f?ts"))
	assert.False(t, IsPattern("news/fts"))
	assert.False(t, IsPattern("fts_quality"))
	assert.False(t, IsPattern(""))
}

func TestJudgmentsPath_StrategyVsExplicitPath(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "demo")
	makeTrack(t, track)
	tr, err := Resolve(Inputs{TrackArg: track})
	require.NoError(t, err)

	t.Run("bare name expands to convention", func(t *testing.T) {
		assert.Equal(t,
			filepath.Join(canonical(t, track), "trec", "annotations.lexical.yaml"),
			tr.JudgmentsPath("lexical"))
	})
	t.Run("explicit path used verbatim", func(t *testing.T) {
		p := "/tmp/some/other.yaml"
		assert.Equal(t, p, tr.JudgmentsPath(p))
	})
	t.Run("empty returns empty", func(t *testing.T) {
		assert.Equal(t, "", tr.JudgmentsPath(""))
	})
}

func TestQrelsPath_StrategySuffix(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "demo")
	makeTrack(t, track)
	tr, _ := Resolve(Inputs{TrackArg: track})
	canon := canonical(t, track)
	assert.Equal(t,
		filepath.Join(canon, "trec", "qrels.claude-api.tsv"),
		tr.QrelsPath("claude-api"))
}

func TestReportPath_UsesRunID(t *testing.T) {
	dir := t.TempDir()
	track := filepath.Join(dir, "tracks", "demo")
	makeTrack(t, track)
	tr, _ := Resolve(Inputs{TrackArg: track})
	canon := canonical(t, track)
	assert.Equal(t,
		filepath.Join(canon, "reports", "2026-05-21T10-00-00-run-abc123.json"),
		tr.ReportPath("2026-05-21T10-00-00-run-abc123"))
	assert.Equal(t,
		filepath.Join(canon, "reports", "latest.json"),
		tr.LatestReportPath())
}
