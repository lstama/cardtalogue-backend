package gen

import (
	_ "embed"
)

// BackendSpec is the pinned Cardtalogue backend contract (OpenAPI 3.x), synced
// from the cardtalogue-api-contract repo by `make sync-contract`. It is embedded
// so the contract-validation middleware can load it at runtime without a
// filesystem dependency. Regenerate openapi.go from this file with `make generate`.
//
//go:embed backend.yaml
var BackendSpec []byte
