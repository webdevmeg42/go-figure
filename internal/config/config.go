package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration to support YAML strings like "5s" and "500ms".
type Duration struct{ time.Duration }

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a string like \"5s\", got %s", value.Tag)
	}
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: must be a Go duration string like \"5s\" or \"500ms\"", value.Value)
	}
	d.Duration = dur
	return nil
}

// Suite is the top-level structure of a go-figure YAML test file.
type Suite struct {
	Tests []Test `yaml:"tests"`
}

// Test represents a single HTTP test case.
type Test struct {
	Name       string     `yaml:"name"`
	Request    Request    `yaml:"request"`
	Assertions Assertions `yaml:"assertions"`
}

// Request describes the HTTP request to send.
type Request struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
	Timeout Duration          `yaml:"timeout"`
}

// Assertions describes what to verify about the HTTP response.
// All fields are optional; omitting a field skips that assertion.
type Assertions struct {
	Status     *int              `yaml:"status"`
	Body       *BodyAssertions   `yaml:"body"`
	Headers    map[string]string `yaml:"headers"`
	LatencyMs  *int              `yaml:"latency_ms"`
	JSONSchema string            `yaml:"json_schema"`
}

// BodyAssertions holds body assertion rules.
type BodyAssertions struct {
	Contains string `yaml:"contains"`
	Exact    string `yaml:"exact"`
}

// Parse reads a YAML test file at path, validates required fields, and returns the Suite.
// Returns an error if the file cannot be read, YAML is invalid, or required
// fields (name, method, url) are missing from any test.
// If request.timeout is omitted, it defaults to 10s.
// Applies {{VAR}} substitution from environment variables before returning.
func Parse(path string) (*Suite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var suite Suite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	for i := range suite.Tests {
		t := &suite.Tests[i]
		if t.Name == "" {
			return nil, fmt.Errorf("test %d: missing name", i)
		}
		if t.Request.Method == "" {
			return nil, fmt.Errorf("test %q: missing request.method", t.Name)
		}
		if t.Request.URL == "" {
			return nil, fmt.Errorf("test %q: missing request.url", t.Name)
		}
		if t.Request.Timeout.Duration == 0 {
			t.Request.Timeout.Duration = 10 * time.Second
		}
	}

	Substitute(&suite)
	return &suite, nil
}
