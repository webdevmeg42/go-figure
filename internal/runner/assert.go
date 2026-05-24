package runner

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/webdevmeg42/go-figure/internal/config"
	"github.com/xeipuuv/gojsonschema"
)

func assertStatus(expected *int, got int) string {
	if expected == nil {
		return ""
	}
	if *expected != got {
		return fmt.Sprintf("status: expected %d, got %d", *expected, got)
	}
	return ""
}

func assertBodyContains(contains, body string) string {
	if contains == "" {
		return ""
	}
	if !strings.Contains(body, contains) {
		return fmt.Sprintf("body: expected to contain %q", contains)
	}
	return ""
}

func assertBodyExact(exact, body string) string {
	if exact == "" {
		return ""
	}
	if exact != body {
		return fmt.Sprintf("body: expected %q, got %q", exact, body)
	}
	return ""
}

func assertHeaders(expected map[string]string, got http.Header) []string {
	var failures []string
	for key, expectedVal := range expected {
		gotVal := got.Get(key)
		if gotVal == "" {
			failures = append(failures, fmt.Sprintf("header %q: expected %q, not present", key, expectedVal))
			continue
		}
		if gotVal != expectedVal {
			failures = append(failures, fmt.Sprintf("header %q: expected %q, got %q", key, expectedVal, gotVal))
		}
	}
	return failures
}

func assertLatency(maxMs *int, actual time.Duration) string {
	if maxMs == nil {
		return ""
	}
	max := time.Duration(*maxMs) * time.Millisecond
	if actual > max {
		return fmt.Sprintf("latency: expected ≤%dms, got %dms", *maxMs, actual.Milliseconds())
	}
	return ""
}

func assertJSONSchema(schema, body string) string {
	if schema == "" {
		return ""
	}
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(body)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Sprintf("json_schema: validation error: %v", err)
	}
	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return fmt.Sprintf("json_schema: %s", strings.Join(errs, "; "))
	}
	return ""
}

// evaluate runs all assertions for a response and returns human-readable failure messages.
// An empty slice means all assertions passed.
func evaluate(
	a config.Assertions,
	statusCode int,
	body string,
	headers http.Header,
	latency time.Duration,
) []string {
	var failures []string

	if msg := assertStatus(a.Status, statusCode); msg != "" {
		failures = append(failures, msg)
	}
	if a.Body != nil {
		if msg := assertBodyContains(a.Body.Contains, body); msg != "" {
			failures = append(failures, msg)
		}
		if msg := assertBodyExact(a.Body.Exact, body); msg != "" {
			failures = append(failures, msg)
		}
	}
	for _, msg := range assertHeaders(a.Headers, headers) {
		failures = append(failures, msg)
	}
	if msg := assertLatency(a.LatencyMs, latency); msg != "" {
		failures = append(failures, msg)
	}
	if msg := assertJSONSchema(a.JSONSchema, body); msg != "" {
		failures = append(failures, msg)
	}
	return failures
}
