package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/webdevmeg42/go-figure/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTemp writes content to a temporary YAML file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestParse_ValidMinimal(t *testing.T) {
	path := writeTemp(t, `
tests:
  - name: "health check"
    request:
      method: GET
      url: http://example.com/health
`)
	suite, err := config.Parse(path)
	require.NoError(t, err)
	require.Len(t, suite.Tests, 1)
	assert.Equal(t, "health check", suite.Tests[0].Name)
	assert.Equal(t, "GET", suite.Tests[0].Request.Method)
	assert.Equal(t, "http://example.com/health", suite.Tests[0].Request.URL)
	assert.Equal(t, 10*time.Second, suite.Tests[0].Request.Timeout.Duration,
		"timeout should default to 10s when omitted")
}

func TestParse_ValidFull(t *testing.T) {
	path := writeTemp(t, `
tests:
  - name: "get user"
    request:
      method: GET
      url: http://example.com/users/1
      timeout: 5s
      headers:
        Authorization: Bearer token123
    assertions:
      status: 200
      body:
        contains: "email"
      headers:
        Content-Type: application/json
      latency_ms: 500
      json_schema: '{"type":"object"}'
`)
	suite, err := config.Parse(path)
	require.NoError(t, err)
	require.Len(t, suite.Tests, 1)

	test := suite.Tests[0]
	assert.Equal(t, 5*time.Second, test.Request.Timeout.Duration)
	assert.Equal(t, "Bearer token123", test.Request.Headers["Authorization"])

	require.NotNil(t, test.Assertions.Status)
	assert.Equal(t, 200, *test.Assertions.Status)

	require.NotNil(t, test.Assertions.Body)
	assert.Equal(t, "email", test.Assertions.Body.Contains)

	require.NotNil(t, test.Assertions.LatencyMs)
	assert.Equal(t, 500, *test.Assertions.LatencyMs)

	assert.Equal(t, `{"type":"object"}`, test.Assertions.JSONSchema)
}

func TestParse_MissingName(t *testing.T) {
	path := writeTemp(t, `
tests:
  - request:
      method: GET
      url: http://example.com
`)
	_, err := config.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing name")
}

func TestParse_MissingMethod(t *testing.T) {
	path := writeTemp(t, `
tests:
  - name: "no method"
    request:
      url: http://example.com
`)
	_, err := config.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing request.method")
}

func TestParse_MissingURL(t *testing.T) {
	path := writeTemp(t, `
tests:
  - name: "no url"
    request:
      method: GET
`)
	_, err := config.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing request.url")
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := config.Parse("/nonexistent/file.yaml")
	require.Error(t, err)
}

func TestParse_FileNotFound_ErrorMessage(t *testing.T) {
	_, err := config.Parse("/nonexistent/file.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}

func TestParse_InvalidYAML(t *testing.T) {
	path := writeTemp(t, "tests:\n  - name: \"bad yaml\"\n    request:\n\tmethod: GET\n")
	_, err := config.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing YAML")
}

func TestParse_InvalidDuration(t *testing.T) {
	path := writeTemp(t, `
tests:
  - name: "bad timeout"
    request:
      method: GET
      url: http://example.com
      timeout: "bananas"
`)
	_, err := config.Parse(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bananas")
}

func TestSubstitute_ReplacesEnvVars(t *testing.T) {
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("AUTH_TOKEN", "secret")

	path := writeTemp(t, `
tests:
  - name: "substitution test"
    request:
      method: GET
      url: "{{BASE_URL}}/users"
      headers:
        Authorization: "Bearer {{AUTH_TOKEN}}"
`)
	suite, err := config.Parse(path)
	require.NoError(t, err)

	test := suite.Tests[0]
	assert.Equal(t, "http://localhost:8080/users", test.Request.URL)
	assert.Equal(t, "Bearer secret", test.Request.Headers["Authorization"])
}

func TestSubstitute_UnsetVarLeftAsIs(t *testing.T) {
	// Ensure the var is not set
	t.Setenv("UNSET_VAR", "")

	path := writeTemp(t, `
tests:
  - name: "unset var"
    request:
      method: GET
      url: "{{UNSET_VAR}}/path"
`)
	suite, err := config.Parse(path)
	require.NoError(t, err)
	// Unset vars are left as-is (not substituted)
	assert.Equal(t, "{{UNSET_VAR}}/path", suite.Tests[0].Request.URL)
}
