package config

import (
	"os"
	"regexp"
)

// placeholderRe matches {{VAR_NAME}} placeholders (uppercase letters, digits, underscores).
var placeholderRe = regexp.MustCompile(`\{\{([A-Z0-9_]+)\}\}`)

// Substitute replaces all {{VAR}} placeholders in suite string fields with the
// corresponding environment variable values. Placeholders for unset or empty
// variables are left unchanged.
func Substitute(suite *Suite) {
	replace := func(s string) string {
		return placeholderRe.ReplaceAllStringFunc(s, func(match string) string {
			key := placeholderRe.FindStringSubmatch(match)[1]
			if val := os.Getenv(key); val != "" {
				return val
			}
			return match
		})
	}

	for i := range suite.Tests {
		t := &suite.Tests[i]
		t.Request.URL = replace(t.Request.URL)
		t.Request.Body = replace(t.Request.Body)
		for k, v := range t.Request.Headers {
			t.Request.Headers[k] = replace(v)
		}
	}
}
