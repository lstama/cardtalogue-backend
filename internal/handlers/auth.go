package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/middleware"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

// ---- Auth policy (per-operation) ----

type authLevel int

const (
	authPublic      authLevel = iota // no auth required
	authJWT                          // valid JWT required
	authEditorAdmin                  // JWT + editor or admin role
	authAdmin                        // JWT + admin role
)

// authLevelFor returns the auth requirement for a contract operation.
// operationID is the capitalized Go method name (e.g. "GetProtectedCards").

func (s *shared) authLevelFor(operationID string) authLevel {
	switch operationID {
	// Public (no auth). Login/Refresh/Logout are public but may return 401.
	case "Health", "Login", "Refresh", "Logout",
		"GetCards", "SearchCards", "GetCardById",
		"GetGames", "SearchGames",
		"GetLanguages", "SearchLanguages",
		"GetPacks", "SearchPacks",
		"GetRarities", "SearchRarities",
		"ScanMedia":
		return authPublic
	// JWT only.
	case "GetStorages", "GetStorageById", "SearchStorages",
		"GetMediaById", "UpdateMedia", "DeleteMedia",
		"GetProtectedCards":
		return authJWT
	// JWT + editor/admin.
	case "CreateCard", "UpdateCard", "UpdateProtectedCard",
		"CreateGame", "UpdateGame",
		"CreateLanguage", "UpdateLanguage",
		"CreatePack", "UpdatePack",
		"CreateRarity", "UpdateRarity",
		"CreateStorage", "UpdateStorage",
		"UploadMedia",
		"GetProtectedCardById":
		return authEditorAdmin
	// JWT + admin.
	case "GetUsers", "GetUserById", "CreateUser", "UpdateUser", "DeleteUser",
		"DeleteGame", "DeleteLanguage", "DeletePack", "DeleteRarity", "DeleteStorage":
		return authAdmin
	default:
		return authJWT
	}
}

// AuthMiddleware returns the per-operation strict middleware. It is exported so
// cmd/api can pass it to gen.NewStrictHandlerWithOptions.

func (s *shared) AuthMiddleware() gen.StrictMiddlewareFunc {
	return s.authMiddleware
}

// authMiddleware wraps a strict handler with per-operation auth. It uses the
// existing gin-based authenticate/hasRole helpers (which write the 401/403
// response directly). Returning (nil, nil) after the response is written lets
// the strict handler serialize nothing.

func (s *shared) authMiddleware(next gen.StrictHandlerFunc, operationID string) gen.StrictHandlerFunc {
	level := s.authLevelFor(operationID)
	if level == authPublic {
		return next
	}
	return func(ctx *gin.Context, request any) (any, error) {
		if !middleware.Authenticate(ctx, s.cfg.JWTSecret, true) {
			return nil, nil // 401 already written by authenticate
		}
		switch level {
		case authAdmin:
			if !middleware.HasRole(ctx, []string{"admin"}) {
				return nil, nil // 403 already written by hasRole
			}
		case authEditorAdmin:
			if !middleware.HasRole(ctx, []string{"admin", "editor"}) {
				return nil, nil // 403 already written by hasRole
			}
		}
		return next(ctx, request)
	}
}

func fromLoginRequest(r *gen.LoginRequest) *models.LoginRequest {
	if r == nil {
		return nil
	}
	return &models.LoginRequest{
		Email:    string(r.Email),
		Password: r.Password,
	}
}

func (s *shared) Login(ctx context.Context, req gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	authResp, err := s.services.Auth.Login(fromLoginRequest(req.Body))
	if err != nil {
		return gen.Login401JSONResponse{UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Error: "invalid credentials"}}, nil
	}
	u := toUserResponse(&authResp.User)
	return gen.Login200JSONResponse{
		AccessToken:  strPtr(authResp.AccessToken),
		RefreshToken: strPtr(authResp.RefreshToken),
		User:         &u,
	}, nil
}

func (s *shared) Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	return gen.Logout200JSONResponse{Message: strPtr("logged out")}, nil
}

func (s *shared) Refresh(ctx context.Context, req gen.RefreshRequestObject) (gen.RefreshResponseObject, error) {
	authResp, err := s.services.Auth.RefreshToken(req.Body.RefreshToken)
	if err != nil {
		return gen.Refresh401JSONResponse{UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Error: "invalid refresh token"}}, nil
	}
	return gen.Refresh200JSONResponse{
		AccessToken: strPtr(authResp.AccessToken),
	}, nil
}
