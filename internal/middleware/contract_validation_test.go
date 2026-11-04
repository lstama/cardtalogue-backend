package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// With validation disabled (the default), the router must not build and the
// middleware must be a no-op passthrough — enabling it never changes behavior.
func TestContractValidationDisabledIsNoop(t *testing.T) {
	t.Setenv(EnvContractValidate, "0")
	if _, ok := newContractRouter(); ok {
		t.Fatal("newContractRouter() should return ok=false when validation is disabled")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ContractValidation())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("disabled middleware must pass through unchanged; got %d", rec.Code)
	}
}

// With validation enabled, the embedded spec must load and build a router.
func TestContractValidationEnabledBuildsRouter(t *testing.T) {
	t.Setenv(EnvContractValidate, "1")
	router, ok := newContractRouter()
	if !ok || router == nil {
		t.Fatal("newContractRouter() should build a router from the embedded spec when enabled")
	}
}

// A request that matches a contract operation with valid parameters passes.
func TestContractValidationPassesValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv(EnvContractValidate, "1")
	r := gin.New()
	r.Use(ContractValidation())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("valid /health request should pass validation; got %d body=%s", rec.Code, rec.Body.String())
	}
}

// A request that matches a contract operation but violates a parameter constraint
// (non-integer :id where the spec requires uint32) is rejected with 400.
func TestContractValidationRejectsInvalidPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv(EnvContractValidate, "1")
	r := gin.New()
	r.Use(ContractValidation())
	r.GET("/cards/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/cards/abc", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-integer id should be rejected with 400; got %d body=%s", rec.Code, rec.Body.String())
	}
}

// A request to a path that is NOT in the contract passes through unchanged, so
// new or unlisted routes are never blocked by validation.
func TestContractValidationPassesUnlistedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv(EnvContractValidate, "1")
	r := gin.New()
	r.Use(ContractValidation())
	r.GET("/api/v1/brand-new", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/brand-new", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("unlisted route should pass through unchanged; got %d body=%s", rec.Code, rec.Body.String())
	}
}
