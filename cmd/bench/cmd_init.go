package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/*.tmpl
var initTemplates embed.FS

type initFlags struct {
	force bool
}

type initContext struct {
	Name           string
	SuiteRel       string
	AnnotationsRel string
}

func newInitCmd() *cobra.Command {
	var f initFlags
	cmd := &cobra.Command{
		Use:   "init <track>",
		Short: "Scaffold a new evaluation track folder",
		Long: `Creates a self-contained evaluation track under tracks/<name>/:

  tracks/<name>/
    spec.yaml          # engines + jobs + defaults
    suite.yaml         # query templates + queries
    trec/              # generated pool, annotations, qrels live here
    reports/           # one JSON per bench run
    README.md          # workflow notes

The folder IS the track — no hidden state, no selector. Run any subcommand
either by name (bench run my_track) or path (bench run --track ./elsewhere).`,
		Example: "  bench init fts_quality_v2",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeInit(cmd, f, args[0])
		},
	}
	cmd.Flags().BoolVar(&f.force, "force", false, "Overwrite existing files in the track folder")
	return cmd
}

func executeInit(cmd *cobra.Command, f initFlags, name string) error {
	if err := validateTrackName(name); err != nil {
		return err
	}

	// Bare name → tracks/<name>/. A path-shaped arg is used verbatim, so power
	// users can drop a track anywhere.
	root := name
	if !pathLike(name) {
		root = filepath.Join("tracks", name)
	}

	if info, err := os.Stat(root); err == nil && info.IsDir() && !f.force {
		entries, _ := os.ReadDir(root)
		if len(entries) > 0 {
			return fmt.Errorf("track %q already exists and is non-empty (use --force to overwrite)", root)
		}
	}
	for _, sub := range []string{"trec", "reports"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Join(root, sub), err)
		}
	}

	ctx := initContext{
		Name:           filepath.Base(root),
		SuiteRel:       filepath.Join(root, "suite.yaml"),
		AnnotationsRel: filepath.Join("trec", "annotations.lexical.yaml"),
	}

	files := []struct {
		tmpl string
		dest string
	}{
		{"templates/spec.yaml.tmpl", filepath.Join(root, "spec.yaml")},
		{"templates/suite.yaml.tmpl", filepath.Join(root, "suite.yaml")},
		{"templates/README.md.tmpl", filepath.Join(root, "README.md")},
	}

	for _, file := range files {
		if err := renderTemplate(file.tmpl, file.dest, ctx, f.force); err != nil {
			return err
		}
	}

	_ = os.WriteFile(filepath.Join(root, "trec", ".gitkeep"), nil, 0644)
	_ = os.WriteFile(filepath.Join(root, "reports", ".gitkeep"), nil, 0644)

	cmd.Printf("Track created: %s/\n", root)
	cmd.Printf("Next: edit %s/suite.yaml, then run:\n", root)
	cmd.Printf("  bench validate %s\n", filepath.Base(root))
	return nil
}

func pathLike(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '/' || s[0] == '.' {
		return true
	}
	for _, r := range s {
		if r == '/' {
			return true
		}
	}
	return false
}

func renderTemplate(srcPath, destPath string, ctx initContext, force bool) error {
	if _, err := os.Stat(destPath); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", destPath)
	}
	raw, err := initTemplates.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read embedded template %s: %w", srcPath, err)
	}
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", srcPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return fmt.Errorf("execute template %s: %w", srcPath, err)
	}
	if err := os.WriteFile(destPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write %s: %w", destPath, err)
	}
	return nil
}

func validateTrackName(name string) error {
	if name == "" {
		return fmt.Errorf("track name is empty")
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-' || r == '_':
		default:
			return fmt.Errorf("track name may only contain [a-zA-Z0-9_-]: %q", name)
		}
	}
	return nil
}
