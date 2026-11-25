package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// newSiteverifyServer returns an httptest server that asserts the CAP
// /siteverify request shape and replies with the given JSON body and status.
// wantPath is the full request path expected, including any site-key prefix.
func newSiteverifyServer(t *testing.T, wantPath, wantSecret, wantResponse, reply string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want none", r.URL.RawQuery)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if payload["secret"] != wantSecret {
			t.Errorf("secret = %q, want %q", payload["secret"], wantSecret)
		}
		if payload["response"] != wantResponse {
			t.Errorf("response = %q, want %q", payload["response"], wantResponse)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
}

func testCapVerifier(t *testing.T, serverURL string) *capVerifier {
	t.Helper()
	return &capVerifier{
		url:    serverURL,
		secret: "test-secret",
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

func TestCapVerifierVerify(t *testing.T) {
	tests := []struct {
		name    string
		reply   string
		status  int
		want    bool
		wantErr bool
	}{
		{name: "success true", reply: `{"success":true}`, status: http.StatusOK, want: true},
		{name: "success false", reply: `{"success":false}`, status: http.StatusOK, want: false},
		{name: "malformed body", reply: `not-json`, status: http.StatusOK, wantErr: true},
		// Current behavior: the HTTP status code is not inspected, only the
		// decoded `success` field. Documented here so a future change is a
		// deliberate one.
		{name: "error status with success false", reply: `{"success":false}`, status: http.StatusInternalServerError, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := newSiteverifyServer(t, "/siteverify", "test-secret", "cap.1.0.0:proof", tc.reply, tc.status)
			defer server.Close()

			got, err := testCapVerifier(t, server.URL).Verify(context.Background(), "cap.1.0.0:proof")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Verify() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Verify() error = %v, want nil", err)
			}
			if got != tc.want {
				t.Errorf("Verify() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCapVerifierVerifyTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close() // no listener left: the request must fail

	got, err := testCapVerifier(t, url).Verify(context.Background(), "cap.1.0.0:proof")
	if err == nil {
		t.Fatalf("Verify() error = nil, want a transport error")
	}
	if got {
		t.Errorf("Verify() = true on error, want false")
	}
}

func TestNoopVerifierDisabled(t *testing.T) {
	var v Verifier = noopVerifier{}
	if v.Enabled() {
		t.Error("noopVerifier.Enabled() = true, want false")
	}
	ok, err := v.Verify(context.Background(), "")
	if err != nil || !ok {
		t.Errorf("noopVerifier.Verify() = (%v, %v), want (true, nil)", ok, err)
	}
}

func TestNewVerifierModeOffIsDisabled(t *testing.T) {
	// appEnv is not "production", so the fail-fast path must not trigger.
	v := NewVerifier("off", "development", "", "")
	if v.Enabled() {
		t.Error("NewVerifier(off).Enabled() = true, want false")
	}
	if _, ok := v.(noopVerifier); !ok {
		t.Errorf("NewVerifier(off) = %T, want noopVerifier", v)
	}
}

func TestNewVerifierModeOnIsEnabled(t *testing.T) {
	v := NewVerifier("on", "production", "https://captcha.test/site-key", "secret")
	if !v.Enabled() {
		t.Error("NewVerifier(on).Enabled() = false, want true")
	}
	cv, ok := v.(*capVerifier)
	if !ok {
		t.Fatalf("NewVerifier(on) = %T, want *capVerifier", v)
	}
	if cv.url != "https://captcha.test/site-key" || cv.secret != "secret" {
		t.Errorf("capVerifier = {url:%q secret:%q}, want the configured values", cv.url, cv.secret)
	}
	if cv.client == nil || cv.client.Timeout == 0 {
		t.Error("capVerifier.client has no timeout, want a bounded timeout")
	}
}

// TestNewVerifierModeTypoStaysEnabled pins the (unchanged) CAPTCHA_MODE
// semantics: only the exact string "off" disables the gate. A typo or an
// unknown value therefore keeps CAP verification enabled — it fails closed
// rather than silently disabling login protection.
func TestNewVerifierModeTypoStaysEnabled(t *testing.T) {
	for _, mode := range []string{"off ", " off", "OFF", "On", "enabled", "yes", ""} {
		t.Run("mode="+mode, func(t *testing.T) {
			v := NewVerifier(mode, "development", "https://captcha.test/site-key", "secret")
			if !v.Enabled() {
				t.Fatalf("NewVerifier(%q).Enabled() = false, want true (unrecognized modes stay enabled)", mode)
			}
			if _, ok := v.(*capVerifier); !ok {
				t.Fatalf("NewVerifier(%q) = %T, want *capVerifier", mode, v)
			}
		})
	}
}

// TestNewCAPVerifierValidation covers the startup validation behind
// NewVerifier: an enabled CAP verifier must have a non-blank secret and an
// absolute http(s) CAP_URL carrying a site-key path segment, otherwise every
// gated request would fail at runtime.
func TestNewCAPVerifierValidation(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		secret   string
		wantURL  string   // normalized CAP_URL for valid configs
		wantErr  string   // substring of the expected error, "" means valid
		notInErr []string // substrings that must NOT leak into the error
	}{
		// -- valid configurations, including normalization --
		{name: "https url", url: "https://captcha.test/site-key", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "http url with port", url: "http://captcha.test:8080/site-key", secret: "secret", wantURL: "http://captcha.test:8080/site-key"},
		{name: "trailing slash trimmed", url: "https://captcha.test/site-key/", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "repeated trailing slashes trimmed", url: "https://captcha.test/site-key///", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "surrounding whitespace trimmed", url: " \thttps://captcha.test/site-key\n ", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "whitespace and trailing slashes trimmed", url: "  https://captcha.test/site-key/  ", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "space before trailing slash trimmed", url: "https://captcha.test/site-key /", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "space and slash mixes trimmed", url: " \thttps://captcha.test/site-key / / \n", secret: "secret", wantURL: "https://captcha.test/site-key"},
		{name: "padded secret trimmed", url: "https://captcha.test/site-key", secret: "  secret\t", wantURL: "https://captcha.test/site-key"},
		// Accepted but NOT rewritten: interior empty path segments and dot
		// segments are an operator error the reviewer advised accepting rather
		// than silently collapsing (a silent rewrite is what hid the
		// trailing-slash bug). Pinned so any change is deliberate.
		{name: "internal duplicate slashes accepted", url: "https://captcha.test//site-key", secret: "secret", wantURL: "https://captcha.test//site-key"},
		{name: "dot segment accepted", url: "https://captcha.test/site-key/..", secret: "secret", wantURL: "https://captcha.test/site-key/.."},

		// -- rejected configurations --
		{name: "empty secret", url: "https://captcha.test/site-key", secret: "", wantErr: "CAP_SECRET is empty"},
		{name: "whitespace-only secret", url: "https://captcha.test/site-key", secret: " \t\n", wantErr: "CAP_SECRET is empty"},
		{name: "empty url", url: "", secret: "secret", wantErr: "CAP_URL is empty"},
		{name: "whitespace-only url", url: " \t\n ", secret: "secret", wantErr: "CAP_URL is empty"},
		{name: "relative url", url: "captcha.test/site-key", secret: "secret", wantErr: "absolute http or https URL"},
		{name: "scheme-less absolute", url: "//captcha.test/site-key", secret: "secret", wantErr: "absolute http or https URL"},
		{name: "unsupported scheme", url: "ftp://captcha.test/site-key", secret: "secret", wantErr: "absolute http or https URL"},
		{name: "hostless url", url: "https://", secret: "secret", wantErr: "must include a host"},
		{name: "unparseable url", url: "https://captcha.test/\x7f", secret: "secret", wantErr: "not a valid URL"},
		{name: "missing site key", url: "https://captcha.test", secret: "secret", wantErr: "site-key path segment"},
		{name: "site key only slash", url: "https://captcha.test/", secret: "secret", wantErr: "site-key path segment"},
		{name: "site key only repeated slashes", url: "https://captcha.test//", secret: "secret", wantErr: "site-key path segment"},
		{name: "site key whitespace only", url: "https://captcha.test/ ", secret: "secret", wantErr: "site-key path segment"},
		{name: "userinfo", url: "https://user:sup3r@captcha.test/site-key", secret: "secret", wantErr: "userinfo", notInErr: []string{"sup3r"}},
		{name: "userinfo with unsupported scheme", url: "ftp://user:sup3r@captcha.test/site-key", secret: "secret", wantErr: "absolute http or https URL", notInErr: []string{"sup3r"}},
		{name: "malformed url with userinfo (parse branch)", url: "https://user:sup3r@captcha.test/%zz", secret: "secret", wantErr: "not a valid URL", notInErr: []string{"sup3r", "captcha.test"}},
		{name: "control character with userinfo (parse branch)", url: "https://user:sup3r@captcha.test/\x7f", secret: "secret", wantErr: "not a valid URL", notInErr: []string{"sup3r", "captcha.test"}},
		{name: "query string", url: "https://captcha.test/site-key?x=1", secret: "secret", wantErr: "query string"},
		{name: "query string with credentials", url: "https://user:sup3r@captcha.test/site-key?x=1", secret: "secret", wantErr: "userinfo", notInErr: []string{"sup3r"}},
		{name: "empty query string", url: "https://captcha.test/site-key?", secret: "secret", wantErr: "query string"},
		{name: "fragment", url: "https://captcha.test/site-key#frag", secret: "secret", wantErr: "fragment"},
		{name: "empty fragment", url: "https://captcha.test/site-key#", secret: "secret", wantErr: "fragment"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := newCAPVerifier(tc.url, tc.secret)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("newCAPVerifier(%q, secret set=%t) error = %v, want nil", tc.url, tc.secret != "", err)
				}
				if v == nil {
					t.Fatalf("newCAPVerifier(%q) = nil, want a verifier", tc.url)
				}
				if v.url != tc.wantURL {
					t.Errorf("normalized CAP_URL = %q, want %q", v.url, tc.wantURL)
				}
				if v.secret != strings.TrimSpace(tc.secret) {
					t.Errorf("CAP_SECRET was not stored trimmed (got %d bytes, want %d)", len(v.secret), len(strings.TrimSpace(tc.secret)))
				}
				if strings.HasSuffix(v.url, "/") {
					t.Errorf("normalized CAP_URL %q still ends with a slash", v.url)
				}
				if v.client == nil || v.client.Timeout == 0 {
					t.Error("capVerifier.client has no timeout, want a bounded timeout")
				}
				return
			}
			if err == nil {
				t.Fatalf("newCAPVerifier(%q) error = nil, want one containing %q", tc.url, tc.wantErr)
			}
			if v != nil {
				t.Errorf("newCAPVerifier returned a verifier alongside the error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
			if tc.notInErr != nil {
				for _, secretish := range tc.notInErr {
					if strings.Contains(err.Error(), secretish) {
						t.Errorf("error leaks %q: %q", secretish, err.Error())
					}
				}
			}
			if strings.Contains(err.Error(), tc.secret) && strings.TrimSpace(tc.secret) != "" {
				t.Errorf("error leaks the CAP secret: %q", err.Error())
			}
		})
	}
}

// TestNewCAPVerifierNormalizedVerifyPath proves the normalized URL is actually
// used for the /siteverify call: whitespace and trailing slashes are gone and
// the site-key segment is preserved.
func TestNewCAPVerifierNormalizedVerifyPath(t *testing.T) {
	server := newSiteverifyServer(t, "/site-key/siteverify", "test-secret", "cap.1.0.0:proof", `{"success":true}`, http.StatusOK)
	defer server.Close()

	v, err := newCAPVerifier("  "+server.URL+"/site-key//  ", "  test-secret\t")
	if err != nil {
		t.Fatalf("newCAPVerifier error = %v, want nil", err)
	}
	ok, err := v.Verify(context.Background(), "cap.1.0.0:proof")
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if !ok {
		t.Error("Verify() = false, want true")
	}
}

// TestNewCAPVerifierSpaceBeforeSlashVerifyPath proves D2's fix at request
// level: "…/site-key /" must post to /site-key/siteverify, not a %20-suffixed
// path.
func TestNewCAPVerifierSpaceBeforeSlashVerifyPath(t *testing.T) {
	server := newSiteverifyServer(t, "/site-key/siteverify", "test-secret", "cap.1.0.0:proof", `{"success":true}`, http.StatusOK)
	defer server.Close()

	v, err := newCAPVerifier("  "+server.URL+"/site-key /  ", "  test-secret\t")
	if err != nil {
		t.Fatalf("newCAPVerifier error = %v, want nil", err)
	}
	if v.url != server.URL+"/site-key" {
		t.Errorf("normalized CAP_URL = %q, want %q", v.url, server.URL+"/site-key")
	}
	ok, err := v.Verify(context.Background(), "cap.1.0.0:proof")
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if !ok {
		t.Error("Verify() = false, want true")
	}
}

// TestIsProduction pins the environment comparison used by the production
// CAPTCHA_MODE=off safeguard: case- and whitespace-insensitive.
func TestIsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{env: "production", want: true},
		{env: "Production", want: true},
		{env: "PRODUCTION", want: true},
		{env: " production ", want: true},
		{env: "\tProduction\n", want: true},
		{env: "development", want: false},
		{env: "prod", want: false},
		{env: "production2", want: false},
		{env: "", want: false},
	}

	for _, tc := range tests {
		if got := isProduction(tc.env); got != tc.want {
			t.Errorf("isProduction(%q) = %v, want %v", tc.env, got, tc.want)
		}
	}
}

// TestVerifyHandlesUnbuildableRequest ensures a malformed CAP_URL surfaces as
// an error from Verify instead of panicking on a nil *http.Request.
func TestVerifyHandlesUnbuildableRequest(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "control character", url: "https://captcha.test/\x7f"},
		{name: "invalid escape", url: "https://captcha.test/%zz"},
		{name: "space in host", url: "https://captcha test/site-key"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &capVerifier{url: tc.url, secret: "secret", client: &http.Client{Timeout: time.Second}}

			got, err := v.Verify(context.Background(), "cap.1.0.0:proof")
			if err == nil {
				t.Fatalf("Verify() with CAP_URL %q error = nil, want a request-building error", tc.url)
			}
			if got {
				t.Errorf("Verify() = true on error, want false")
			}
		})
	}
}

// TestNewVerifierFailsFastOnMisconfiguration re-executes the test binary so the
// log.Fatalf guards can be observed as a process exit instead of killing the
// run. Valid configurations are included as controls that must exit 0, so an
// unconditionally fatal constructor cannot make this test pass.
func TestNewVerifierFailsFastOnMisconfiguration(t *testing.T) {
	if os.Getenv("CAPTCHA_HELPER_PROCESS") == "1" {
		// Child process: NewVerifier either returns (exit 0) or log.Fatalf's.
		NewVerifier(
			os.Getenv("CAPTCHA_HELPER_MODE"),
			os.Getenv("CAPTCHA_HELPER_ENV"),
			os.Getenv("CAPTCHA_HELPER_URL"),
			os.Getenv("CAPTCHA_HELPER_SECRET"),
		)
		return // exit 0: the guard did not stop startup
	}

	tests := []struct {
		name         string
		mode         string
		env          string
		url          string
		secret       string
		wantExitZero bool
		wantOutput   string   // substring expected in the child output, "" if none
		notInOutput  []string // substrings that must not leak into the child output
	}{
		{
			name: "malformed url", mode: "on", env: "development",
			url: "captcha.test/site-key", secret: "secret",
			wantOutput: "misconfigured",
		},
		{
			name: "empty url", mode: "on", env: "development",
			url: "", secret: "secret",
			wantOutput: "misconfigured",
		},
		{
			name: "missing secret", mode: "on", env: "development",
			url: "https://captcha.test/site-key", secret: "",
			wantOutput: "misconfigured",
		},
		{
			name: "url missing site key", mode: "on", env: "development",
			url: "https://captcha.test/", secret: "secret",
			wantOutput: "site-key path segment",
		},
		{
			// A malformed URL can only carry userinfo if it is rejected in the
			// parse branch, so this pins the no-credential-in-fatal-log rule.
			name: "malformed url with userinfo does not leak the password", mode: "on", env: "development",
			url: "https://user:sup3r@captcha.test/%zz", secret: "secret",
			wantOutput:  "misconfigured",
			notInOutput: []string{"sup3r", "captcha.test"},
		},
		{
			name: "production off is fatal", mode: "off", env: "production",
			url: "", secret: "",
			wantOutput: "not allowed in production",
		},
		{
			name: "production off mixed case and whitespace is fatal", mode: "off", env: " Production ",
			url: "", secret: "",
			wantOutput: "not allowed in production",
		},
		{
			name: "valid config control exits zero", mode: "on", env: "development",
			url: "https://captcha.test/site-key", secret: "control-secret",
			wantExitZero: true,
			notInOutput:  []string{"misconfigured", "control-secret"},
		},
		{
			name: "development off control exits zero", mode: "off", env: "development",
			url: "", secret: "",
			wantExitZero: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestNewVerifierFailsFastOnMisconfiguration$")
			cmd.Env = append(os.Environ(),
				"CAPTCHA_HELPER_PROCESS=1",
				"CAPTCHA_HELPER_MODE="+tc.mode,
				"CAPTCHA_HELPER_ENV="+tc.env,
				"CAPTCHA_HELPER_URL="+tc.url,
				"CAPTCHA_HELPER_SECRET="+tc.secret,
			)
			out, err := cmd.CombinedOutput()

			if tc.wantExitZero {
				if err != nil {
					t.Fatalf("child exited with %v, want exit 0; output:\n%s", err, out)
				}
			} else {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("child err = %v, want a non-zero exit; output:\n%s", err, out)
				}
				if exitErr.ExitCode() == 0 {
					t.Fatalf("child exit code = 0, want non-zero; output:\n%s", out)
				}
			}

			if tc.wantOutput != "" && !strings.Contains(string(out), tc.wantOutput) {
				t.Errorf("child output does not contain %q:\n%s", tc.wantOutput, out)
			}
			for _, secretish := range tc.notInOutput {
				if strings.Contains(string(out), secretish) {
					t.Errorf("child output leaks %q:\n%s", secretish, out)
				}
			}
		})
	}
}
