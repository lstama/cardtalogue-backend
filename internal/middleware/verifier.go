package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Verifier interface {
	Enabled() bool
	Verify(ctx context.Context, token string) (bool, error)
}

func NewVerifier(captchaMode string, appEnv string, capUrl string, capSecret string) Verifier {
	if captchaMode == "off" {
		if appEnv == "production" {
			log.Fatal("CAPTCHA_MODE=off is not allowed in production")
		}
		log.Println("WARNING: captcha disabled")
		return noopVerifier{}
	}
	return &capVerifier{
		url:    capUrl,
		secret: capSecret,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type capVerifier struct {
	url, secret string // url includes the site key: https://cap.tama.id/<site-key>
	client      *http.Client
}

func (v *capVerifier) Enabled() bool { return true }

func (v *capVerifier) Verify(ctx context.Context, token string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"secret": v.secret, "response": token})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, v.url+"/siteverify", bytes.NewReader(body))
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
