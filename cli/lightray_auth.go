package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	lightrayDeviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"
	lightrayOAuthScopes         = "openid profile email"
)

type lightrayDiscoveryDocument struct {
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
}

type lightrayDeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type lightrayTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

type lightrayOAuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
	StatusCode  int    `json:"-"`
}

func (e *lightrayOAuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("oauth error %s: %s", e.Code, e.Description)
	}
	if e.Code != "" {
		return fmt.Sprintf("oauth error %s", e.Code)
	}
	return fmt.Sprintf("oauth request failed with HTTP %d", e.StatusCode)
}

var errLightrayAuthenticationAborted = errors.New("Lightray authentication aborted")

type lightrayAuthenticator struct {
	httpClient    *http.Client
	output        io.Writer
	keyboardInput io.ReadCloser
	keyboardAbort bool
	sleep         func(context.Context, time.Duration) error
	now           func() time.Time
}

func defaultSleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a lightrayAuthenticator) client() *http.Client {
	if a.httpClient != nil {
		return a.httpClient
	}
	return http.DefaultClient
}

func (a lightrayAuthenticator) writer() io.Writer {
	if a.output != nil {
		return a.output
	}
	return io.Discard
}

func (a lightrayAuthenticator) sleepFunc() func(context.Context, time.Duration) error {
	if a.sleep != nil {
		return a.sleep
	}
	return defaultSleep
}

func (a lightrayAuthenticator) nowFunc() func() time.Time {
	if a.now != nil {
		return a.now
	}
	return time.Now
}

func authenticateWithLightray(ctx context.Context, lightrayURL, clientID string, output io.Writer) (string, error) {
	return lightrayAuthenticator{output: output, keyboardAbort: true}.Authenticate(ctx, lightrayURL, clientID)
}

func (a lightrayAuthenticator) Authenticate(ctx context.Context, lightrayURL, clientID string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	baseURL, err := normalizeLightrayBaseURL(lightrayURL)
	if err != nil {
		return "", err
	}

	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		clientID = defaultLightrayClientID
	}

	discovery, err := a.discover(ctx, baseURL)
	if err != nil {
		return "", err
	}

	deviceAuth, err := a.startDeviceAuthorization(ctx, discovery.DeviceAuthorizationEndpoint, clientID)
	if err != nil {
		return "", err
	}

	a.printDeviceInstructions(deviceAuth)

	pollCtx, stopKeyboardAbort, abortCh := a.withKeyboardAbort(ctx)
	defer stopKeyboardAbort()

	accessToken, err := a.pollForToken(pollCtx, discovery.TokenEndpoint, clientID, deviceAuth)
	if err != nil {
		if abortErr := readKeyboardAbortError(abortCh); abortErr != nil {
			return "", abortErr
		}
		return "", err
	}

	return accessToken, nil
}

func normalizeLightrayBaseURL(raw string) (string, error) {
	u, err := parseBaseURL(raw, "Lightray URL")
	if err != nil {
		return "", err
	}

	path := strings.TrimRight(u.Path, "/")
	if path == "/api" || strings.HasSuffix(path, "/api") {
		path = strings.TrimSuffix(path, "/api")
	}
	u.Path = path
	u.RawPath = ""

	return strings.TrimRight(u.String(), "/"), nil
}

func lightrayAPIURL(raw string) (string, error) {
	u, err := parseBaseURL(raw, "Lightray URL")
	if err != nil {
		return "", err
	}

	path := strings.TrimRight(u.Path, "/")
	if path == "" {
		path = "/api"
	} else if !strings.HasSuffix(path, "/api") {
		path += "/api"
	}
	u.Path = path
	u.RawPath = ""

	return u.String(), nil
}

func parseBaseURL(raw, label string) (*url.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("%s is required", label)
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", label, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid %s: absolute URL with scheme and host is required", label)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid %s: URL scheme must be http or https", label)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("invalid %s: query strings and fragments are not supported", label)
	}

	return u, nil
}

func lightrayURLJoin(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

func (a lightrayAuthenticator) discover(ctx context.Context, baseURL string) (*lightrayDiscoveryDocument, error) {
	discoveryURL := lightrayURLJoin(baseURL, "/.well-known/openid-configuration")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Lightray OpenID configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Lightray OpenID configuration: HTTP %d%s", resp.StatusCode, responseSnippetSuffix(resp.Body))
	}

	var discovery lightrayDiscoveryDocument
	if err := decodeLightrayJSONResponse(resp, &discovery, "Lightray OpenID configuration"); err != nil {
		return nil, err
	}
	if discovery.DeviceAuthorizationEndpoint == "" {
		return nil, fmt.Errorf("Lightray OpenID configuration does not advertise device_authorization_endpoint")
	}
	if discovery.TokenEndpoint == "" {
		return nil, fmt.Errorf("Lightray OpenID configuration does not advertise token_endpoint")
	}

	return &discovery, nil
}

func (a lightrayAuthenticator) startDeviceAuthorization(ctx context.Context, endpoint, clientID string) (*lightrayDeviceAuthorizationResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", lightrayOAuthScopes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to start Lightray device authorization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to start Lightray device authorization: %w", parseOAuthError(resp))
	}

	var deviceAuth lightrayDeviceAuthorizationResponse
	if err := decodeLightrayJSONResponse(resp, &deviceAuth, "Lightray device authorization response"); err != nil {
		return nil, err
	}
	if deviceAuth.DeviceCode == "" {
		return nil, fmt.Errorf("Lightray device authorization response did not include device_code")
	}
	if deviceAuth.UserCode == "" {
		return nil, fmt.Errorf("Lightray device authorization response did not include user_code")
	}
	if deviceAuth.Interval <= 0 {
		deviceAuth.Interval = 5
	}
	if deviceAuth.ExpiresIn <= 0 {
		deviceAuth.ExpiresIn = 10 * 60
	}

	return &deviceAuth, nil
}

func (a lightrayAuthenticator) printDeviceInstructions(deviceAuth *lightrayDeviceAuthorizationResponse) {
	verificationURL := deviceAuth.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = deviceAuth.VerificationURI
	}

	out := a.writer()
	fmt.Fprintln(out, "Open this URL to authenticate with Lightray:")
	if verificationURL != "" {
		fmt.Fprintln(out, verificationURL)
	} else {
		fmt.Fprintln(out, "(verification URL was not provided by the server)")
	}
	fmt.Fprintf(out, "\nCode: %s\n", deviceAuth.UserCode)
	fmt.Fprintln(out, "Press Ctrl+C or Ctrl+D to abort.")
	fmt.Fprintln(out, "Waiting for approval...")
}

func (a lightrayAuthenticator) withKeyboardAbort(ctx context.Context) (context.Context, func(), <-chan error) {
	input := a.keyboardInput
	if input == nil && a.keyboardAbort {
		openedInput, err := openKeyboardAbortInput()
		if err != nil {
			return ctx, func() {}, nil
		}
		input = openedInput
	}
	if input == nil {
		return ctx, func() {}, nil
	}

	abortCtx, cancel := context.WithCancel(ctx)
	abortCh := make(chan error, 1)
	done := make(chan struct{})
	var stopOnce sync.Once

	stop := func() {
		stopOnce.Do(func() {
			close(done)
			_ = input.Close()
			cancel()
		})
	}

	go func() {
		reader := bufio.NewReader(input)
		for {
			_, err := reader.ReadString('\n')

			select {
			case <-done:
				return
			default:
			}

			if errors.Is(err, io.EOF) {
				abortCh <- errLightrayAuthenticationAborted
				cancel()
				return
			}
			if err != nil {
				return
			}
		}
	}()

	return abortCtx, stop, abortCh
}

func openKeyboardAbortInput() (io.ReadCloser, error) {
	return os.OpenFile("/dev/tty", os.O_RDONLY, 0)
}

func readKeyboardAbortError(abortCh <-chan error) error {
	if abortCh == nil {
		return nil
	}

	select {
	case err := <-abortCh:
		return err
	default:
		return nil
	}
}

func (a lightrayAuthenticator) pollForToken(ctx context.Context, endpoint, clientID string, deviceAuth *lightrayDeviceAuthorizationResponse) (string, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	now := a.nowFunc()
	expiresAt := now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)
	sleep := a.sleepFunc()

	for {
		accessToken, err := a.requestToken(ctx, endpoint, clientID, deviceAuth.DeviceCode)
		if err == nil {
			fmt.Fprintln(a.writer(), "Lightray authentication complete.")
			return accessToken, nil
		}

		var oauthErr *lightrayOAuthError
		if !errors.As(err, &oauthErr) {
			return "", err
		}

		switch oauthErr.Code {
		case "authorization_pending":
			// Continue polling after the server-provided interval.
		case "slow_down":
			interval += 5 * time.Second
		case "access_denied":
			return "", fmt.Errorf("Lightray device authorization was denied")
		case "expired_token":
			return "", fmt.Errorf("Lightray device authorization expired")
		default:
			return "", oauthErr
		}

		if !now().Add(interval).Before(expiresAt) {
			return "", fmt.Errorf("Lightray device authorization expired")
		}

		if err := sleep(ctx, interval); err != nil {
			return "", err
		}
	}
}

func (a lightrayAuthenticator) requestToken(ctx context.Context, endpoint, clientID, deviceCode string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", lightrayDeviceCodeGrantType)
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to poll Lightray token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", parseOAuthError(resp)
	}

	var tokenResponse lightrayTokenResponse
	if err := decodeLightrayJSONResponse(resp, &tokenResponse, "Lightray token response"); err != nil {
		return "", err
	}
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("Lightray token response did not include access_token")
	}
	if tokenResponse.TokenType != "" && !strings.EqualFold(tokenResponse.TokenType, "Bearer") {
		return "", fmt.Errorf("Lightray token response used unsupported token_type %q", tokenResponse.TokenType)
	}

	return tokenResponse.AccessToken, nil
}

func decodeLightrayJSONResponse(resp *http.Response, target any, description string) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", description, err)
	}

	trimmed := strings.TrimSpace(string(body))
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") || strings.HasPrefix(trimmed, "<") {
		return fmt.Errorf("%s returned HTML instead of JSON; this usually means the Lightray OAuth/device auth routes are not deployed, or the URL is not the Lightray origin", description)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse %s: %w", description, err)
	}

	return nil
}

func parseOAuthError(resp *http.Response) error {
	var oauthErr lightrayOAuthError
	oauthErr.StatusCode = resp.StatusCode

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if readErr != nil {
		return fmt.Errorf("HTTP %d: failed to read error response: %w", resp.StatusCode, readErr)
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &oauthErr); err == nil && oauthErr.Code != "" {
			return &oauthErr
		}
	}

	snippet := strings.TrimSpace(string(body))
	if snippet != "" {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
	}

	return &oauthErr
}

func responseSnippetSuffix(reader io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(reader, 4096))
	if err != nil {
		return ""
	}

	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return ""
	}

	return ": " + snippet
}
