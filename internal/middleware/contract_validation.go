package middleware

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

// EnvContractValidate is the env var that enables the contract-validation
// middleware. It defaults to OFF, so wiring this middleware in never changes
// default runtime behavior: when unset, ContractValidation is a no-op passthrough.
const EnvContractValidate = "API_CONTRACT_VALIDATE"

// ContractValidation validates each incoming request against the pinned backend
// contract. Requests that do not match any contract operation pass through
// unchanged, so new or unlisted routes are never blocked. It is a no-op unless
// EnvContractValidate == "1".
//
// NOTE: this validates request structure (params, query, body) against the spec.
// It does NOT enforce auth schemes; authentication stays with the existing
// middleware. Validation is opt-in precisely because the contract may be stricter
// than current behavior — enable it deliberately once discrepancies are resolved.
func ContractValidation() gin.HandlerFunc {
	router, ok := newContractRouter()
	if !ok {
		// Spec failed to load/validate, or validation disabled: never crash.
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			// Not in the contract — pass through unchanged.
			c.Next()
			return
		}
		in := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				MultiError:         true,
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		if err := openapi3filter.ValidateRequest(c.Request.Context(), in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contract validation failed: " + err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

// apiPrefix is the router mount point the app uses for every contract route
// (all routes are grouped under /api/v1 in cmd/api/main.go). The contract paths
// are relative to the API root (e.g. /cards/{id}), so the middleware rewrites the
// spec's server path to this prefix before matching request paths.
var apiPrefix = "/api/v1"

// newContractRouter loads the embedded spec, makes its server path equal to the
// app's API mount point so path matching works, and builds a kin-openapi router.
// It returns ok=false when validation is disabled or the spec is invalid.
func newContractRouter() (routers.Router, bool) {
	if os.Getenv(EnvContractValidate) != "1" {
		return nil, false
	}
	ctx := context.Background()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(gen.BackendSpec)
	if err != nil {
		log.Printf("contract validation: spec load failed, disabling: %v", err)
		return nil, false
	}
	// The spec's servers carry a host (e.g. http://localhost:8080) and paths are
	// relative to the API root; point the server at the app's mount prefix so
	// FindRoute matches the full request path (e.g. /api/v1/cards/123).
	doc.Servers = []*openapi3.Server{{URL: apiPrefix}}
	if err := doc.Validate(ctx); err != nil {
		log.Printf("contract validation: spec validation failed, disabling: %v", err)
		return nil, false
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		log.Printf("contract validation: router build failed, disabling: %v", err)
		return nil, false
	}
	return router, true
}
