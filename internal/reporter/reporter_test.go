package reporter_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/webdevmeg42/go-figure/internal/reporter"
	"github.com/webdevmeg42/go-figure/internal/runner"
	"github.com/stretchr/testify/assert"
)

// Disable ANSI color codes so we can assert on plain text.
func TestMain(m *testing.M) {
	color.NoColor = true
	os.Exit(m.Run())
}

func TestPrint_AllPass(t *testing.T) {
	results := []runner.Result{
		{TestName: "health check", Passed: true, Duration: 100 * time.Millisecond},
		{TestName: "get user", Passed: true, Duration: 200 * time.Millisecond},
	}

	var buf bytes.Buffer
	reporter.Print(&buf, results)
	out := buf.String()

	assert.Contains(t, out, "✓")
	assert.Contains(t, out, "health check")
	assert.Contains(t, out, "get user")
	assert.Contains(t, out, "2/2 passed")
	assert.NotContains(t, out, "failed")
}

func TestPrint_WithFailures(t *testing.T) {
	results := []runner.Result{
		{TestName: "health check", Passed: true, Duration: 100 * time.Millisecond},
		{
			TestName: "create user",
			Passed:   false,
			Duration: 80 * time.Millisecond,
			Failures: []string{
				"status: expected 201, got 400",
				`body: expected to contain "id"`,
			},
		},
	}

	var buf bytes.Buffer
	reporter.Print(&buf, results)
	out := buf.String()

	assert.Contains(t, out, "✓")
	assert.Contains(t, out, "✗")
	assert.Contains(t, out, "create user")
	assert.Contains(t, out, "status: expected 201, got 400")
	assert.Contains(t, out, `body: expected to contain "id"`)
	assert.Contains(t, out, "1/2 passed")
	assert.Contains(t, out, "1 failed")
}

func TestPrint_SummaryTiming(t *testing.T) {
	results := []runner.Result{
		{TestName: "a", Passed: true, Duration: 300 * time.Millisecond},
		{TestName: "b", Passed: true, Duration: 700 * time.Millisecond},
	}

	var buf bytes.Buffer
	reporter.Print(&buf, results)
	out := buf.String()

	// Total should be ~1000ms
	assert.True(t,
		strings.Contains(out, "1000ms") || strings.Contains(out, "999ms") || strings.Contains(out, "1001ms"),
		"summary should contain total duration near 1000ms, got: %s", out)
}
