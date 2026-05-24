package runner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/webdevmeg42/go-figure/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- assertStatus ---

func TestAssertStatus_Pass(t *testing.T) {
	code := 200
	assert.Empty(t, assertStatus(&code, 200))
}

func TestAssertStatus_Fail(t *testing.T) {
	code := 201
	msg := assertStatus(&code, 400)
	assert.Equal(t, "status: expected 201, got 400", msg)
}

func TestAssertStatus_Nil(t *testing.T) {
	assert.Empty(t, assertStatus(nil, 200), "nil status should skip assertion")
}

// --- assertBodyContains ---

func TestAssertBodyContains_Pass(t *testing.T) {
	assert.Empty(t, assertBodyContains("email", `{"email":"a@b.com"}`))
}

func TestAssertBodyContains_Fail(t *testing.T) {
	msg := assertBodyContains("email", `{"name":"Ada"}`)
	assert.Equal(t, `body: expected to contain "email"`, msg)
}

func TestAssertBodyContains_Empty(t *testing.T) {
	assert.Empty(t, assertBodyContains("", `{"any":"body"}`), "empty contains should skip")
}

// --- assertBodyExact ---

func TestAssertBodyExact_Pass(t *testing.T) {
	assert.Empty(t, assertBodyExact("hello", "hello"))
}

func TestAssertBodyExact_Fail(t *testing.T) {
	msg := assertBodyExact("hello", "world")
	assert.Contains(t, msg, "expected")
	assert.Contains(t, msg, "hello")
}

func TestAssertBodyExact_Empty(t *testing.T) {
	assert.Empty(t, assertBodyExact("", "anything"), "empty exact should skip")
}

// --- assertHeaders ---

func TestAssertHeaders_Pass(t *testing.T) {
	expected := map[string]string{"Content-Type": "application/json"}
	got := http.Header{"Content-Type": []string{"application/json"}}
	assert.Empty(t, assertHeaders(expected, got))
}

func TestAssertHeaders_Fail_WrongValue(t *testing.T) {
	expected := map[string]string{"Content-Type": "application/json"}
	got := http.Header{"Content-Type": []string{"text/html"}}
	failures := assertHeaders(expected, got)
	assert.Len(t, failures, 1)
	assert.Contains(t, failures[0], "Content-Type")
}

func TestAssertHeaders_Fail_Missing(t *testing.T) {
	expected := map[string]string{"X-Request-ID": "abc"}
	got := http.Header{}
	failures := assertHeaders(expected, got)
	assert.Len(t, failures, 1)
	assert.Contains(t, failures[0], "X-Request-ID")
}

func TestAssertHeaders_Nil(t *testing.T) {
	assert.Empty(t, assertHeaders(nil, http.Header{}))
}

// --- assertLatency ---

func TestAssertLatency_Pass(t *testing.T) {
	max := 500
	assert.Empty(t, assertLatency(&max, 200*time.Millisecond))
}

func TestAssertLatency_Fail(t *testing.T) {
	max := 100
	msg := assertLatency(&max, 250*time.Millisecond)
	assert.Contains(t, msg, "100ms")
	assert.Contains(t, msg, "250ms")
}

func TestAssertLatency_Nil(t *testing.T) {
	assert.Empty(t, assertLatency(nil, 999*time.Millisecond), "nil latency should skip")
}

// --- assertJSONSchema ---

func TestAssertJSONSchema_Pass(t *testing.T) {
	schema := `{"type":"object","required":["id"]}`
	body := `{"id":1,"name":"Ada"}`
	assert.Empty(t, assertJSONSchema(schema, body))
}

func TestAssertJSONSchema_Fail_MissingRequired(t *testing.T) {
	schema := `{"type":"object","required":["id"]}`
	body := `{"name":"Ada"}`
	msg := assertJSONSchema(schema, body)
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "json_schema")
}

func TestAssertJSONSchema_Empty(t *testing.T) {
	assert.Empty(t, assertJSONSchema("", `{"any":"body"}`), "empty schema should skip")
}

// --- Runner integration tests ---

func statusPtr(n int) *int { return &n }

func TestRun_AllPass(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"email":"a@b.com"}`))
	}))
	defer srv.Close()

	suite := &config.Suite{
		Tests: []config.Test{
			{
				Name: "get user",
				Request: config.Request{
					Method:  "GET",
					URL:     srv.URL + "/users/1",
					Timeout: config.Duration{Duration: 5 * time.Second},
				},
				Assertions: config.Assertions{
					Status: statusPtr(200),
					Body:   &config.BodyAssertions{Contains: "email"},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				},
			},
		},
	}

	results := Run(suite)
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
	assert.Empty(t, results[0].Failures)
	assert.Equal(t, "get user", results[0].TestName)
	assert.Positive(t, results[0].Duration)
}

func TestRun_StatusFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	suite := &config.Suite{
		Tests: []config.Test{
			{
				Name: "expect 200",
				Request: config.Request{
					Method:  "GET",
					URL:     srv.URL,
					Timeout: config.Duration{Duration: 5 * time.Second},
				},
				Assertions: config.Assertions{
					Status: statusPtr(200),
				},
			},
		},
	}

	results := Run(suite)
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	require.Len(t, results[0].Failures, 1)
	assert.Equal(t, "status: expected 200, got 400", results[0].Failures[0])
}

func TestRun_BodyContainsFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Ada"}`))
	}))
	defer srv.Close()

	suite := &config.Suite{
		Tests: []config.Test{
			{
				Name: "body contains email",
				Request: config.Request{
					Method:  "GET",
					URL:     srv.URL,
					Timeout: config.Duration{Duration: 5 * time.Second},
				},
				Assertions: config.Assertions{
					Body: &config.BodyAssertions{Contains: "email"},
				},
			},
		},
	}

	results := Run(suite)
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Contains(t, results[0].Failures[0], `"email"`)
}

func TestRun_NoAssertions_Passes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	suite := &config.Suite{
		Tests: []config.Test{
			{
				Name: "smoke test",
				Request: config.Request{
					Method:  "GET",
					URL:     srv.URL,
					Timeout: config.Duration{Duration: 5 * time.Second},
				},
			},
		},
	}

	results := Run(suite)
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed, "test with no assertions should pass if request succeeds")
}

func TestRun_PostWithBody(t *testing.T) {
	var receivedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":42}`))
	}))
	defer srv.Close()

	suite := &config.Suite{
		Tests: []config.Test{
			{
				Name: "create user",
				Request: config.Request{
					Method:  "POST",
					URL:     srv.URL + "/users",
					Body:    `{"name":"Ada"}`,
					Headers: map[string]string{"Content-Type": "application/json"},
					Timeout: config.Duration{Duration: 5 * time.Second},
				},
				Assertions: config.Assertions{
					Status: statusPtr(201),
					Body:   &config.BodyAssertions{Contains: "id"},
				},
			},
		},
	}

	results := Run(suite)
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
	assert.Equal(t, `{"name":"Ada"}`, string(receivedBody))
}
