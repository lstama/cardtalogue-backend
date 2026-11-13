
## 2025-XX-XX — Phase 3 model/conversion reconciliation (continuation)
- Verified current state: models for cards, packs, users ALREADY match contract DTOs
  (extended in earlier commit fad8ed2). Summary's "domain mismatch" note is stale.
- Only `media` still diverges: contract MediaResponse has Type/FileName/FileThumbURL/
  FileSize/IsDeleted/IsPending; model Media has none (model has extra Provider).
- Decision: extend model Media with Type/FileName/FileThumbURL/FileSize/IsDeleted/
  IsPending so conversions are clean field-by-field (per user: "extend models to match
  contract"). Migrations use gorm AutoMigrate -> new columns auto-created.
- strict.go is stale (written against old Title/BackText schema). Will rewrite conversions
  to match current models.
- middleware.authenticate/hasRole are unexported -> rename to Authenticate/HasRole so
  package `handlers` can call them.
