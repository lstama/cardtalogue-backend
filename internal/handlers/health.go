package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

// ---- Handlers ----

func (s *shared) Health(ctx context.Context, req gen.HealthRequestObject) (gen.HealthResponseObject, error) {
	return &gen.Health200JSONResponse{Status: strPtr("healthy")}, nil
}
