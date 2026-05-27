package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
)

// ─── color palette ────────────────────────────────────────────────────────────

var (
	cOK   = color.New(color.FgGreen, color.Bold)
	cFail = color.New(color.FgRed, color.Bold)
	cWarn = color.New(color.FgYellow)
	cDim  = color.New(color.FgHiBlack)
	cBold = color.New(color.Bold)
	cCyan = color.New(color.FgCyan)
)

// ─── one-line status printers ─────────────────────────────────────────────────

func printDone(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", cOK.Sprint("✓"), msg)
}

func printWarn(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", cWarn.Sprint("⚠"), msg)
}

// ─── validate status coloring ─────────────────────────────────────────────────

func colorStatus(s string) string {
	switch s {
	case "OK":
		return cOK.Sprint(s)
	case "INVALID", "TEMPLATE_ERR":
		return cFail.Sprint(s)
	case "SKIP":
		return cDim.Sprint(s)
	case "UNSUPPORTED":
		return cWarn.Sprint(s)
	default:
		return s
	}
}

// ─── spinner ──────────────────────────────────────────────────────────────────

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner is a braille-dot progress indicator that writes to stderr.
// It is a no-op when stderr is not a TTY (CI, piped output).
type Spinner struct {
	stop chan struct{}
	done chan struct{}
}

// startSpinner starts an animated spinner with the given status message and
// returns a handle. Always call Stop() — it is safe even on error paths.
func startSpinner(suffix string) *Spinner {
	s := &Spinner{stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(s.done)
		if !stderrIsTTY() {
			// Not a TTY: print a static "running…" line once then wait.
			fmt.Fprintf(os.Stderr, "%s  %s\n", cCyan.Sprint("…"), suffix)
			<-s.stop
			return
		}
		for i := 0; ; {
			fmt.Fprintf(os.Stderr, "\r%s  %s",
				cCyan.Sprint(spinFrames[i%len(spinFrames)]), suffix)
			select {
			case <-s.stop:
				fmt.Fprint(os.Stderr, "\r\033[K") // erase spinner line
				return
			case <-time.After(80 * time.Millisecond):
				i++
			}
		}
	}()
	return s
}

// Stop halts the spinner and clears the line. Blocks until the goroutine exits.
func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
}

// stderrIsTTY reports whether os.Stderr is an interactive terminal.
func stderrIsTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
