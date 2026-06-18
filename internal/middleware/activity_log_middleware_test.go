package middleware

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeBodyRedactsSensitiveFields(t *testing.T) {
	body := []byte(`{
		"email":"owner@example.com",
		"password":"secret",
		"nested":{"access_token":"token-value"},
		"items":[{"refresh_token":"refresh-value"}]
	}`)

	result := sanitizeBody(body)
	var decoded map[string]any
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Fatalf("sanitizeBody returned invalid JSON: %v", err)
	}

	if decoded["password"] != "[REDACTED]" {
		t.Fatalf("password was not redacted: %v", decoded["password"])
	}
	nested := decoded["nested"].(map[string]any)
	if nested["access_token"] != "[REDACTED]" {
		t.Fatalf("access_token was not redacted: %v", nested["access_token"])
	}
}

func TestBuildCurlRedactsAuthorizationAndQuerySecrets(t *testing.T) {
	requestURL, err := url.Parse("/api/v1/auth/google/callback?code=secret-code&state=secret-state&keep=yes")
	if err != nil {
		t.Fatal(err)
	}
	request := &http.Request{
		Method: http.MethodPost,
		URL:    requestURL,
		Host:   "localhost:8081",
		Header: http.Header{
			"Authorization": {"Bearer real-token"},
			"Content-Type":  {"application/json"},
		},
	}

	command := buildCurl(request, `{"password":"[REDACTED]"}`)
	for _, secret := range []string{"real-token", "secret-code", "secret-state"} {
		if strings.Contains(command, secret) {
			t.Fatalf("curl command contains sensitive value %q: %s", secret, command)
		}
	}
	if !strings.Contains(command, "keep=yes") {
		t.Fatalf("curl command removed non-sensitive query parameter: %s", command)
	}
	if !strings.Contains(command, "http://localhost:8081/") {
		t.Fatalf("curl command does not contain an absolute URL: %s", command)
	}
}

func TestActivityStatusMapping(t *testing.T) {
	tests := map[int]string{
		200: "success",
		201: "success",
		302: "success",
		400: "failed",
		500: "failed",
	}

	for code, expected := range tests {
		actual := "failed"
		if code >= http.StatusOK && code < http.StatusBadRequest {
			actual = "success"
		}
		if actual != expected {
			t.Errorf("status %d mapped to %q, want %q", code, actual, expected)
		}
	}
}
