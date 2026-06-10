# go-figure

> A lightweight Go CLI for running HTTP API tests from a YAML config — with colored output and a CI-friendly exit code.

[![CI](https://github.com/webdevmeg42/go-figure/actions/workflows/ci.yml/badge.svg)](https://github.com/webdevmeg42/go-figure/actions/workflows/ci.yml)

---

## Why I built this

Most HTTP testing tools are either language-specific test frameworks or heavy GUI tools. I wanted something that lives at the CLI level: write a YAML file describing your API expectations, run one command, and get a clear pass/fail result you can drop into any CI pipeline. Go is an ideal fit here — fast startup, single binary, strong standard library for HTTP. This project was also a deliberate exercise in Go's layered package design and using `httptest` for deterministic integration testing without mocking frameworks.

---

## Sample output

```
go-figure v0.1.0  ·  tests.yaml

  ✓  Get post                                213ms
  ✗  Create post                              89ms
       status: expected 201, got 400
       body:   expected to contain "id"
  ✓  List posts                              302ms

  ───────────────────────────────────
  2/3 passed  ·  1 failed  ·  604ms total
```

Exit code `0` when all tests pass, `1` when any fail — works as a CI step without additional tooling.

---

## Installation

**Go install (requires Go 1.22+):**
```bash
go install github.com/webdevmeg42/go-figure@latest
```

**Build from source:**
```bash
git clone https://github.com/webdevmeg42/go-figure.git
cd go-figure
go build -o go-figure .
```

---

## Usage

```bash
go-figure run tests.yaml
```

With environment variables:
```bash
BASE_URL=https://api.example.com AUTH_TOKEN=mytoken go-figure run tests.yaml
```

In GitHub Actions:
```yaml
- name: Run API tests
  run: go-figure run tests.yaml
  env:
    BASE_URL: ${{ secrets.API_BASE_URL }}
    AUTH_TOKEN: ${{ secrets.AUTH_TOKEN }}
```

---

## Test file format

```yaml
tests:
  - name: "Get user profile"
    request:
      method: GET
      url: "{{BASE_URL}}/users/1"
      headers:
        Authorization: "Bearer {{AUTH_TOKEN}}"
      timeout: 5s
    assertions:
      status: 200
      body:
        contains: "email"
      headers:
        Content-Type: "application/json"
      latency_ms: 500
      json_schema: |
        {
          "type": "object",
          "required": ["id", "email"]
        }
```

---

## Assertion reference

| Field | Type | Description |
|-------|------|-------------|
| `assertions.status` | int | HTTP status code must equal this value |
| `assertions.body.contains` | string | Response body must contain this substring |
| `assertions.body.exact` | string | Response body must equal this string exactly |
| `assertions.headers` | map | Each key/value must be present in the response headers |
| `assertions.latency_ms` | int | Response time must be ≤ this value in milliseconds |
| `assertions.json_schema` | string | Response body must conform to this inline JSON schema |

All assertion fields are optional. A test with no assertions runs as a smoke/reachability check.

`request.timeout` defaults to `10s` if omitted.

---

## Environment variables

Any `{{PLACEHOLDER}}` in a YAML string field is substituted from the matching environment variable at runtime. Placeholders must be uppercase letters, digits, and underscores (e.g. `{{BASE_URL}}`, `{{AUTH_TOKEN}}`). Unset variables are left as-is.

This works natively with CI secret injection — no extra tooling needed.

---

## Development

```bash
go test ./... -race    # run all tests
golangci-lint run      # lint
go build -o go-figure  # build binary
```

Tests use `net/http/httptest` — no external services required.
