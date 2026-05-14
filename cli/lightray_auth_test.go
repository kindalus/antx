package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLightrayAPIURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "origin", raw: "https://vcrm.lightray.cloud", want: "https://vcrm.lightray.cloud/api"},
		{name: "trailing slash", raw: "https://vcrm.lightray.cloud/", want: "https://vcrm.lightray.cloud/api"},
		{name: "path", raw: "https://vcrm.lightray.cloud/workspace", want: "https://vcrm.lightray.cloud/workspace/api"},
		{name: "already api", raw: "https://vcrm.lightray.cloud/api", want: "https://vcrm.lightray.cloud/api"},
		{name: "already api with slash", raw: "https://vcrm.lightray.cloud/api/", want: "https://vcrm.lightray.cloud/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lightrayAPIURL(tt.raw)
			if err != nil {
				t.Fatalf("lightrayAPIURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("lightrayAPIURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeLightrayBaseURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "origin", raw: "https://vcrm.lightray.cloud/", want: "https://vcrm.lightray.cloud"},
		{name: "path", raw: "https://vcrm.lightray.cloud/workspace/", want: "https://vcrm.lightray.cloud/workspace"},
		{name: "api suffix", raw: "https://vcrm.lightray.cloud/api", want: "https://vcrm.lightray.cloud"},
		{name: "path api suffix", raw: "https://vcrm.lightray.cloud/workspace/api/", want: "https://vcrm.lightray.cloud/workspace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeLightrayBaseURL(tt.raw)
			if err != nil {
				t.Fatalf("normalizeLightrayBaseURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeLightrayBaseURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestLightrayAuthenticatorDeviceFlow(t *testing.T) {
	var tokenPolls int
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]string{
				"device_authorization_endpoint": server.URL + "/oauth/device_authorization",
				"token_endpoint":                server.URL + "/oauth/token",
			})
		case "/oauth/device_authorization":
			assertFormValue(t, r, "client_id", "terminal-cli")
			assertFormValue(t, r, "scope", lightrayOAuthScopes)
			writeJSON(t, w, map[string]any{
				"device_code":               "device-code-123",
				"user_code":                 "ABCD-EFGH",
				"verification_uri":          server.URL + "/l/device",
				"verification_uri_complete": server.URL + "/l/device?user_code=ABCD-EFGH",
				"expires_in":                60,
				"interval":                  1,
			})
		case "/oauth/token":
			tokenPolls++
			assertFormValue(t, r, "grant_type", lightrayDeviceCodeGrantType)
			assertFormValue(t, r, "client_id", "terminal-cli")
			assertFormValue(t, r, "device_code", "device-code-123")
			switch tokenPolls {
			case 1:
				writeJSONStatus(t, w, http.StatusBadRequest, map[string]string{"error": "authorization_pending"})
			case 2:
				writeJSONStatus(t, w, http.StatusBadRequest, map[string]string{"error": "slow_down"})
			default:
				writeJSON(t, w, map[string]any{
					"access_token": "lightray-access-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
					"scope":        lightrayOAuthScopes,
				})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var output bytes.Buffer
	var sleeps []time.Duration
	authenticator := lightrayAuthenticator{
		httpClient: server.Client(),
		output:     &output,
		sleep: func(ctx context.Context, duration time.Duration) error {
			sleeps = append(sleeps, duration)
			return nil
		},
	}

	token, err := authenticator.Authenticate(context.Background(), server.URL, "terminal-cli")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if token != "lightray-access-token" {
		t.Fatalf("Authenticate returned token %q, want %q", token, "lightray-access-token")
	}
	if tokenPolls != 3 {
		t.Fatalf("token endpoint was polled %d times, want 3", tokenPolls)
	}
	if want := []time.Duration{time.Second, 6 * time.Second}; !reflect.DeepEqual(sleeps, want) {
		t.Fatalf("sleep durations = %v, want %v", sleeps, want)
	}

	text := output.String()
	for _, want := range []string{
		"Open this URL to authenticate with Lightray:",
		server.URL + "/l/device?user_code=ABCD-EFGH",
		"Code: ABCD-EFGH",
		"Waiting for approval...",
		"Lightray authentication complete.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output did not contain %q:\n%s", want, text)
		}
	}
}

func TestNewClientForStartOptionsUsesLightrayAPIURLAndToken(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]string{
				"device_authorization_endpoint": server.URL + "/oauth/device_authorization",
				"token_endpoint":                server.URL + "/oauth/token",
			})
		case "/oauth/device_authorization":
			writeJSON(t, w, map[string]any{
				"device_code":               "device-code-123",
				"user_code":                 "ABCD-EFGH",
				"verification_uri_complete": server.URL + "/l/device?user_code=ABCD-EFGH",
				"expires_in":                60,
				"interval":                  1,
			})
		case "/oauth/token":
			writeJSON(t, w, map[string]any{
				"access_token": "lightray-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case "/api/login/me":
			if got := r.Header.Get("Authorization"); got != "Bearer lightray-access-token" {
				t.Fatalf("Authorization header = %q, want bearer token", got)
			}
			writeJSON(t, w, map[string]any{
				"email":  "test@example.com",
				"name":   "Test User",
				"groups": []string{"testers"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := newClientForStartOptions(context.Background(), StartOptions{
		ServerURL:    server.URL,
		AuthLightray: true,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("newClientForStartOptions returned error: %v", err)
	}

	user, err := client.GetCurrentUser()
	if err != nil {
		t.Fatalf("GetCurrentUser returned error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("user email = %q, want test@example.com", user.Email)
	}
}

func TestStartOptionsValidateLightrayConflicts(t *testing.T) {
	tests := []StartOptions{
		{ServerURL: "https://vcrm.lightray.cloud", AuthLightray: true, APIKey: "key"},
		{ServerURL: "https://vcrm.lightray.cloud", AuthLightray: true, Root: "password"},
		{ServerURL: "https://vcrm.lightray.cloud", AuthLightray: true, JWT: "jwt"},
	}

	for _, options := range tests {
		if err := options.Validate(); err == nil {
			t.Fatalf("Validate(%+v) returned nil, want conflict error", options)
		}
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	writeJSONStatus(t, w, http.StatusOK, value)
}

func writeJSONStatus(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("failed to write JSON response: %v", err)
	}
}

func assertFormValue(t *testing.T, r *http.Request, key, want string) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("%s %s used method %s, want POST", r.URL.Path, key, r.Method)
	}
	if contentType := r.Header.Get("Content-Type"); !strings.Contains(contentType, "application/x-www-form-urlencoded") {
		t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", contentType)
	}
	body, err := ioReadAllAndRestore(r)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("failed to parse request body %q: %v", string(body), err)
	}
	if got := values.Get(key); got != want {
		t.Fatalf("form value %s = %q, want %q (body %s)", key, got, want, string(body))
	}
}

func ioReadAllAndRestore(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func TestParseBaseURLRejectsRelativeURLs(t *testing.T) {
	if _, err := lightrayAPIURL("vcrm.lightray.cloud"); err == nil {
		t.Fatal("lightrayAPIURL accepted a relative URL")
	}
}

func TestLightrayOAuthErrorFormatting(t *testing.T) {
	err := (&lightrayOAuthError{Code: "invalid_client", Description: "Cliente OAuth desconhecido.", StatusCode: 400}).Error()
	want := "oauth error invalid_client: Cliente OAuth desconhecido."
	if err != want {
		t.Fatalf("error string = %q, want %q", err, want)
	}
}
