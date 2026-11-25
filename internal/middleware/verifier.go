package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Verifier interface {
	Enabled() bool
	Verify(ctx context.Context, token string) (bool, error)
}

func NewVerifier(captchaMode string, appEnv string, capUrl string, capSecret string) Verifier {
	if captchaMode == "off" {
		if isProduction(appEnv) {
			log.Fatal("CAPTCHA_MODE=off is not allowed in production")
		}
		log.Println("WARNING: captcha disabled")
		return noopVerifier{}
	}
	v, err := newCAPVerifier(capUrl, capSecret)
	if err != nil {
		// Fail fast: a misconfigured CAP endpoint would otherwise make every
		// gated request (login, media scan) fail ambiguously at runtime.
		log.Fatalf("CAPTCHA_MODE=%q enables CAP verification but it is misconfigured: %v", captchaMode, err)
	}
	return v
}

// isProduction reports whether appEnv names the production environment,
// ignoring surrounding whitespace and letter case (" Production " matches).
func isProduction(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "production")
}

// newCAPVerifier validates the CAP configuration and builds the enabled
// verifier. Verification is enabled for any CAPTCHA_MODE value other than
// "off", so an unrecognized mode still fails closed; it must not run with a
// CAP endpoint that cannot be called.
//
// CAP_URL is the CAP site URL including the site key
// (https://<captcha-host>/<site-key>). Surrounding whitespace and trailing
// slashes are removed so the request is composed as <CAP_URL>/siteverify
// without a doubled slash. A non-empty site-key path segment is required, and
// userinfo/query/fragment components are rejected: userinfo would be echoed by
// a fatal misconfiguration log, and a query or fragment would make the
// /siteverify suffix land in the wrong URL component.
func newCAPVerifier(capURL, capSecret string) (*capVerifier, error) {
	secret := strings.TrimSpace(capSecret)
	if secret == "" {
		return nil, fmt.Errorf("CAP_SECRET is empty; set the CAP server secret")
	}

	siteURL := normalizeCAPURL(capURL)
	if siteURL == "" {
		return nil, fmt.Errorf("CAP_URL is empty; set it to the CAP site URL (https://<captcha-host>/<site-key>)")
	}
	u, err := url.Parse(siteURL)
	if err != nil {
		// Deliberately not wrapping err and not interpolating the raw URL: the
		// returned *url.Error embeds the original string (which may still carry
		// userinfo credentials at this point) and this message reaches a fatal
		// startup log. The message stays static and actionable instead.
		return nil, fmt.Errorf("CAP_URL is not a valid URL; use an absolute http(s) URL of the form https://<captcha-host>/<site-key>")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("CAP_URL %q must be an absolute http or https URL", redactURL(siteURL))
	}
	if u.Host == "" {
		return nil, fmt.Errorf("CAP_URL %q must include a host", redactURL(siteURL))
	}
	if u.User != nil {
		return nil, fmt.Errorf("CAP_URL must not contain userinfo credentials; remove any user:password@ prefix")
	}
	if u.RawQuery != "" || u.ForceQuery {
		return nil, fmt.Errorf("CAP_URL %q must not include a query string", redactURL(siteURL))
	}
	if u.Fragment != "" || strings.Contains(siteURL, "#") {
		return nil, fmt.Errorf("CAP_URL %q must not include a fragment", redactURL(siteURL))
	}
	if strings.Trim(u.Path, "/") == "" {
		return nil, fmt.Errorf("CAP_URL %q must include a site-key path segment (https://<captcha-host>/<site-key>)", redactURL(siteURL))
	}

	return &capVerifier{
		url:    siteURL,
		secret: secret,
		client: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// normalizeCAPURL trims surrounding whitespace and trailing slashes until the
// value stops changing. A single pass is not enough: trimming slashes from
// "https://host/key /" leaves the space that preceded the slash
// ("https://host/key "), which parses to a %20-suffixed path and posts to a
// wrong /siteverify URL (silent 404/outage). Repeating until stable also
// handles mixes such as "https://host/key/ /". Interior empty path segments
// and dot segments are deliberately left untouched (see verifier_test.go).
func normalizeCAPURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	for {
		next := strings.TrimRight(strings.TrimSpace(trimmed), "/")
		if next == trimmed {
			return next
		}
		trimmed = next
	}
}

// redactURL strips userinfo from a URL before it is quoted in an error message,
// so credentials can never reach a log line.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = nil
	return u.String()
}

type capVerifier struct {
	url, secret string // url includes the site key: https://captcha.example.com/<site-key>
	client      *http.Client
}

func (v *capVerifier) Enabled() bool { return true }

func (v *capVerifier) Verify(ctx context.Context, token string) (bool, error) {
	body, err := json.Marshal(map[string]string{"secret": v.secret, "response": token})
	if err != nil {
		return false, fmt.Errorf("encode siteverify request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.url+"/siteverify", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("build siteverify request for %q: %w", v.url+"/siteverify", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := v.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var r struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return false, err
	}
	return r.Success, nil
}

type noopVerifier struct{}

func (noopVerifier) Enabled() bool { return false }

func (noopVerifier) Verify(context.Context, string) (bool, error) { return true, nil }

func RequireCaptcha(v Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !v.Enabled() {
			c.Next()
			return
		}
		token := c.GetHeader("X-Cap-Token")
		if token == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "captcha required"})
			return
		}
		ok, err := v.Verify(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(502, gin.H{"error": "captcha unavailable"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "captcha failed"})
			return
		}
		c.Next()
	}
}
