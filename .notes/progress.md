# Progress

## Phase 3: Contract-First Wiring (COMPLETE)
Goal: wire backend to generated strict server + per-op auth.

### Done
- [x] Investigated models/services/generated code.
- [x] Confirmed card/pack/user models already match contract (extended in fad8ed2).
- [x] Confirmed only `media` diverged (contract MediaResponse has Type/FileName/FileThumbURL/
      FileSize/IsDeleted/IsPending; model Media lacked them).
- [x] Refactored middleware/auth.go -> exposed Authenticate/HasRole (was authenticate/hasRole).
- [x] Added UpdateMedia/DeleteMedia to MediaService interface + impl.
- [x] Wrote strict.go (Services/Shared, 52 handlers, conversions).
- [x] Extend model Media with Type/FileName/FileThumbURL/FileSize/IsDeleted/IsPending.
- [x] Rewrite strict.go conversions to match current models.
- [x] Wire NewShared + NewStrictHandlerWithOptions into cmd/api/main.go (per-op auth).
- [x] go build/vet + gofmt.

### Wiring (final)
- main.go: `shared := handlers.NewShared(services, cfg)`;
  `ginStrict := gen.NewStrictHandlerWithOptions(shared, []gen.StrictMiddlewareFunc{shared.AuthMiddleware()}, gen.StrictGinServerOptions{})`;
  `gen.RegisterHandlersWithOptions(api, ginStrict, gen.GinServerOptions{BaseURL: ""})` on the
  `/api/v1` group. All 52 contract routes auto-registered; no manual route table.
- `shared.AuthMiddleware()` (exported) returns the per-op `StrictMiddlewareFunc`; unexported
  `authMiddleware` does the work (authPublic -> passthrough; else Authenticate + role check).
- Captcha: moved from a gin middleware into `shared.ScanMedia` (ctx is *gin.Context at runtime;
  type-assert, verify X-Cap-Token via shared.verifier). Contract ScanMedia has no 500/403, so
  captcha failures map to 400.
- `/media/:filename` file-serving route kept at root (outside /api/v1) to avoid param conflict
  with contract `/api/v1/media/:id`.
- Contract validation stays opt-in (API_CONTRACT_VALIDATE=1), applied globally as no-op passthrough.

### Verified
- `go build ./...` OK; `go vet ./...` clean; gofmt clean.

### Blocked
- none.

### Key facts
- Auth service: internal/service/auth_service.go (Login/RefreshToken/Logout/ValidateToken/
  ChangePassword). Errors ErrInvalid/ErrExpired/ErrMissing. JWTClaims{UserID,Email,Role,
  TokenType}.
- Contract types: Limit=int32, Offset=int32, Id=uint32, LoginJSONRequestBody=LoginRequest,
  RefreshJSONRequestBody=RefreshTokenRequest.
- /metrics stays root, no auth.
- Migrations use gorm AutoMigrate -> adding model fields auto-creates columns.

## deadcode / game_handler.go dead-code check (2025)
- User asked: reason for `games.go` vs `games_handler.go`? Check with `deadcode`.
- Correction: there is NO `games_handler.go`. Old file is `game_handler.go` (singular).
- Installed `deadcode` (golang.org/x/tools/cmd/deadcode@v0.49.0 -> ~/go/bin).
- deadcode v0.49.0 PANICS (nil ptr in go/callgraph/rta.visitFunc) on this repo under go1.27.3.
  Unmaintained tool; RTA incompatible with go1.27. Tried -test, ./..., v0.21.0 (telemetry build
  error under go1.27), v0.15/v0.10 (no cmd/deadcode pkg). All fail. deadcode unusable here.
- Reachability via grep (conclusive): `GameHandler`/`NewGameHandler` + all 6 methods referenced
  ONLY inside game_handler.go itself. No callers anywhere, no test refs. main.go wires the NEW
  strict handler (handlers.NewShared -> gen.NewStrictHandlerWithOptions -> RegisterHandlersWithOptions),
  NOT NewGameHandler.
- VERDICT: game_handler.go is entirely DEAD code (leftover pre-contract refactor). Delete-safe.
- Only old handler still live: MediaHandler (handlers.NewMediaHandler in main.go:131), serves
  /media/:filename outside contract by design.

## Media handler refactor + test suite (2025)
- Deleted dead pre-contract handlers: auth_handler.go, card_handler.go, game_handler.go,
  language_handler.go, pack_handler.go, rarity_handler.go, storage_handler.go, user_handler.go.
- Removed dead NewStrictHandler from shared.go (setupRouter inlined its logic) + now-unused gen import.
- Refactored media handler to codebase pattern: created media_serve.go folding UploadImage,
  ScanImage, ServeMedia, isValidFilename into methods on *shared. Removed media_handler.go + NewMediaHandler.
- main.go: dropped NewMediaHandler call + unused mediaHandler param from setupRouter; wired legacy
  outside-contract routes via shared: GET /media/:filename, POST /media/upload, POST /media/scan.
  No API break (same paths/methods).
- Test suite: internal/handlers/media_serve_test.go (white-box, httptest + mockMediaService).
  Tests: TestServeMedia (valid file, missing file=404), TestUploadImage (success+scan, missing field=400,
  service err=500), TestScanImage (success, missing field=400, service err=500), TestIsValidFilename.
- mockMediaService embeds service.MediaService + defines UploadImage/ScanImage delegating to uploadFn/scanFn
  fields (must define the methods explicitly — embedding alone dispatches to the nil embedded interface).
- Test expectation notes: gin rejects encoded traversal (..%2f) at 404 before handler; isValidFilename
  guards traversal only, NOT extensions (noext/plain are valid).
- Verify: go build ./... = 0, go vet ./... = 0, go test ./... = ok (handlers + middleware), gofmt -l = empty.

## Mapper file rename (2025)
- Switched mapper files from concern-first `mappers_<domain>.go` to domain-first `<domain>_mappers.go`
  for locality (handler + its mapper adjacent in ls). Renamed 8 files:
  mappers_{cards,games,languages,media,packs,rarities,storages,users}.go -> *_mappers.go.
- Pure rename (no content change); files were untracked so plain mv. Build/vet/test/gofmt all green.
