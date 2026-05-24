package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/webdevmeg42/go-figure/internal/config"
)

// Result holds the outcome of a single test execution.
type Result struct {
	TestName string
	Passed   bool
	Duration time.Duration
	Failures []string // human-readable assertion failure messages
}

// Run executes all tests in the suite sequentially and returns one Result per test.
func Run(suite *config.Suite) []Result {
	results := make([]Result, 0, len(suite.Tests))
	for _, test := range suite.Tests {
		results = append(results, runTest(test))
	}
	return results
}

func runTest(test config.Test) Result {
	timeout := test.Request.Timeout.Duration
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var bodyReader io.Reader = http.NoBody
	if test.Request.Body != "" {
		bodyReader = strings.NewReader(test.Request.Body)
	}

	req, err := http.NewRequestWithContext(ctx, test.Request.Method, test.Request.URL, bodyReader)
	if err != nil {
		return Result{
			TestName: test.Name,
			Passed:   false,
			Failures: []string{fmt.Sprintf("building request: %v", err)},
		}
	}
	for k, v := range test.Request.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return Result{
			TestName: test.Name,
			Passed:   false,
			Duration: latency,
			Failures: []string{fmt.Sprintf("http: %v", err)},
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{
			TestName: test.Name,
			Passed:   false,
			Duration: latency,
			Failures: []string{fmt.Sprintf("reading body: %v", err)},
		}
	}

	failures := evaluate(test.Assertions, resp.StatusCode, string(bodyBytes), resp.Header, latency)
	return Result{
		TestName: test.Name,
		Passed:   len(failures) == 0,
		Duration: latency,
		Failures: failures,
	}
}
