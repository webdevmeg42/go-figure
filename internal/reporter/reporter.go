package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/webdevmeg42/go-figure/internal/runner"
)

var (
	green = color.New(color.FgGreen)
	red   = color.New(color.FgRed)
	dim   = color.New(color.FgHiBlack)
)

// Print renders test results to w with colored pass/fail indicators and a summary line.
// Failure details are indented beneath each failing test.
// Set color.NoColor = true in tests to disable ANSI codes for plain-text assertions.
func Print(w io.Writer, results []runner.Result) {
	fmt.Fprintln(w)

	passed := 0
	var total time.Duration

	for _, r := range results {
		total += r.Duration
		if r.Passed {
			passed++
			green.Fprintf(w, "  ✓  %-36s %dms\n", r.TestName, r.Duration.Milliseconds())
		} else {
			red.Fprintf(w, "  ✗  %-36s %dms\n", r.TestName, r.Duration.Milliseconds())
			for _, f := range r.Failures {
				dim.Fprintf(w, "       %s\n", f)
			}
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "  ───────────────────────────────────")

	failed := len(results) - passed
	summary := fmt.Sprintf("  %d/%d passed", passed, len(results))
	if failed > 0 {
		summary += fmt.Sprintf("  ·  %d failed", failed)
	}
	summary += fmt.Sprintf("  ·  %dms total", total.Milliseconds())

	if failed > 0 {
		red.Fprintln(w, summary)
	} else {
		green.Fprintln(w, summary)
	}
	fmt.Fprintln(w)
}
