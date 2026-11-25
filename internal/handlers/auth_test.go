package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"
)

// fakeVerifier is a middleware.Verifier that records the tokens it is asked to
// verify. It appends to a shared order log so tests can assert that CAPTCHA
// verification happens before credential verification.
type fakeVerifier struct {
	enabled bool
	ok      bool
	err     error
	tokens  []string
	order   *[]string
}

func (v *fakeVerifier) Enabled() bool { return v.enabled }

func (v *fakeVerifier) Verify(_ context.Context, token string) (bool, error) {
	v.tokens = append(v.tokens, token)
	if v.order != nil {
		*v.order = append(*v.order, "verify:"+token)
	}
	return v.ok, v.err
}

// fakeAuthService stubs service.AuthService so a CAPTCHA-blocked login can be
// proven never to reach credential verification.
type fakeAuthService struct {
	service.AuthService
	loginCalls int
	lastReq    *models.LoginRequest
	loginErr   error
	order      *[]string
}

func (a *fakeAuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	a.loginCalls++
	a.lastReq = req
	if a.order != nil {
		*a.order = append(*a.order, "login")
	}
	if a.loginErr != nil {
		return nil, a.loginErr
	}
	return &models.LoginResponse{
		User: models.UserResponse{
			ID:    7,
			Name:  "Tester",
			Email: req.Email,
			Role:  "admin",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil
}

func newCaptchaShared(v *fakeVerifier, a *fakeAuthService) *shared {
	return &shared{
		services: &Services{Auth: a},
		cfg:      &config.Config{CaptchaMode: "on"},
		verifier: v,
	}
}

func loginRequest(token *string) gen.LoginRequestObject {
	body := &gen.LoginJSONRequestBody{Email: "user@example.com", Password: "secret"}
	obj := gen.LoginRequestObject{Body: body}
	if token != nil {
		capToken := gen.CapToken(*token)
		obj.Params.XCapToken = &capToken
	}
	return obj
}

// loginStatus renders the response object through the generated visitor so the
// assertion covers the HTTP status code, not just the Go type.
func loginStatus(t *testing.T, resp gen.LoginResponseObject) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	if err := resp.VisitLoginResponse(w); err != nil {
		t.Fatalf("VisitLoginResponse: %v", err)
	}
	return w
}

func TestLoginCaptchaDisabledAllowsLoginWithoutToken(t *testing.T) {
	v := &fakeVerifier{enabled: false, ok: true}
	a := &fakeAuthService{}
	s := newCaptchaShared(v, a)

	resp, err := s.Login(context.Background(), loginRequest(nil))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if len(v.tokens) != 0 {
		t.Errorf("verifier.Verify called %d times, want 0", len(v.tokens))
	}
	if a.loginCalls != 1 {
		t.Fatalf("AuthService.Login called %d times, want 1", a.loginCalls)
	}
	if got := loginStatus(t, resp).Code; got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}
}

func TestLoginCaptchaDisabledIgnoresToken(t *testing.T) {
	v := &fakeVerifier{enabled: false, ok: false, err: errors.New("should not be called")}
	a := &fakeAuthService{}
	s := newCaptchaShared(v, a)

	resp, err := s.Login(context.Background(), loginRequest(strPtr("ignored-token")))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if len(v.tokens) != 0 {
		t.Errorf("verifier.Verify called %d times, want 0", len(v.tokens))
	}
	if got := loginStatus(t, resp).Code; got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}
}

func TestLoginCaptchaFailures(t *testing.T) {
	token := strPtr("cap.1.0.0:proof")

	tests := []struct {
		name       string
		token      *string
		verifyOK   bool
		verifyErr  error
		wantMsg    string
		wantVerify int
	}{
		{
			name:       "missing token",
			token:      nil,
			wantMsg:    "captcha required",
			wantVerify: 0,
		},
		{
			name:       "empty token",
			token:      strPtr(""),
			wantMsg:    "captcha required",
			wantVerify: 0,
		},
		{
			name:       "verifier error",
			token:      token,
			verifyErr:  errors.New("cap server unreachable"),
			wantMsg:    "captcha unavailable",
			wantVerify: 1,
		},
		{
			name:       "token rejected",
			token:      token,
			verifyOK:   false,
			wantMsg:    "captcha failed",
			wantVerify: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &fakeVerifier{enabled: true, ok: tc.verifyOK, err: tc.verifyErr}
			a := &fakeAuthService{}
			s := newCaptchaShared(v, a)

			resp, err := s.Login(context.Background(), loginRequest(tc.token))
			if err != nil {
				t.Fatalf("Login: %v", err)
			}

			bad, ok := resp.(gen.Login400JSONResponse)
			if !ok {
				t.Fatalf("response type = %T, want gen.Login400JSONResponse", resp)
			}
			if bad.Error != tc.wantMsg {
				t.Errorf("error = %q, want %q", bad.Error, tc.wantMsg)
			}
			if got := loginStatus(t, resp).Code; got != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", got, http.StatusBadRequest)
			}
			if a.loginCalls != 0 {
				t.Errorf("AuthService.Login called %d times, want 0", a.loginCalls)
			}
			if len(v.tokens) != tc.wantVerify {
				t.Errorf("verifier.Verify called %d times, want %d", len(v.tokens), tc.wantVerify)
			}
		})
	}
}

func TestLoginCaptchaValidTokenVerifiesThenAuthenticates(t *testing.T) {
	const token = "cap.1.0.0:proof-token"
	var order []string
	v := &fakeVerifier{enabled: true, ok: true, order: &order}
	a := &fakeAuthService{order: &order}
	s := newCaptchaShared(v, a)

	resp, err := s.Login(context.Background(), loginRequest(strPtr(token)))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if len(v.tokens) != 1 || v.tokens[0] != token {
		t.Fatalf("Verify tokens = %v, want exactly [%q]", v.tokens, token)
	}
	if len(order) != 2 || order[0] != "verify:"+token || order[1] != "login" {
		t.Errorf("call order = %v, want [verify:%s login]", order, token)
	}
	if a.lastReq == nil || a.lastReq.Email != "user@example.com" || a.lastReq.Password != "secret" {
		t.Errorf("AuthService.Login request = %+v, want the decoded credentials", a.lastReq)
	}

	ok, isOK := resp.(gen.Login200JSONResponse)
	if !isOK {
		t.Fatalf("response type = %T, want gen.Login200JSONResponse", resp)
	}
	if ok.Data == nil || ok.Data.AccessToken == nil || *ok.Data.AccessToken != "access-token" {
		t.Errorf("access token = %+v, want access-token", ok.Data)
	}
	if got := loginStatus(t, resp).Code; got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}
}

func TestLoginCaptchaStillReturns401ForBadCredentials(t *testing.T) {
	v := &fakeVerifier{enabled: true, ok: true}
	a := &fakeAuthService{loginErr: errors.New("invalid credentials")}
	s := newCaptchaShared(v, a)

	resp, err := s.Login(context.Background(), loginRequest(strPtr("cap.1.0.0:proof")))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if _, ok := resp.(gen.Login401JSONResponse); !ok {
		t.Fatalf("response type = %T, want gen.Login401JSONResponse", resp)
	}
	if got := loginStatus(t, resp).Code; got != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}

// TestLoginCaptchaEndToEnd wires the generated strict router, so it covers the
// whole path: X-Cap-Token header binding in the generated wrapper ->
// req.Params.XCapToken -> CAPTCHA gate -> AuthService.
func TestLoginCaptchaEndToEnd(t *testing.T) {
	const token = "cap.1.0.0:proof"
	const body = `{"email":"user@example.com","password":"secret"}`

	newRouter := func(v *fakeVerifier, a *fakeAuthService) *gin.Engine {
		s := newCaptchaShared(v, a)
		gin.SetMode(gin.TestMode)
		r := gin.New()
		strict := gen.NewStrictHandlerWithOptions(s, []gen.StrictMiddlewareFunc{s.AuthMiddleware()}, gen.StrictGinServerOptions{})
		gen.RegisterHandlersWithOptions(r.Group("/api/v1"), strict, gen.GinServerOptions{BaseURL: ""})
		return r
	}

	post := func(r *gin.Engine, capToken string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if capToken != "" {
			req.Header.Set("X-Cap-Token", capToken)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("enabled rejects missing header", func(t *testing.T) {
		v := &fakeVerifier{enabled: true, ok: true}
		a := &fakeAuthService{}
		w := post(newRouter(v, a), "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
		if !strings.Contains(w.Body.String(), "captcha required") {
			t.Errorf("body = %s, want captcha required", w.Body.String())
		}
		if a.loginCalls != 0 || len(v.tokens) != 0 {
			t.Errorf("login=%d verify=%d, want 0 and 0", a.loginCalls, len(v.tokens))
		}
	})

	t.Run("enabled forwards the header value to Verify", func(t *testing.T) {
		v := &fakeVerifier{enabled: true, ok: true}
		a := &fakeAuthService{}
		w := post(newRouter(v, a), token)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
		}
		if len(v.tokens) != 1 || v.tokens[0] != token {
			t.Errorf("Verify tokens = %v, want [%q]", v.tokens, token)
		}
		if a.loginCalls != 1 {
			t.Errorf("AuthService.Login called %d times, want 1", a.loginCalls)
		}
	})

	t.Run("disabled lets a token-less login through", func(t *testing.T) {
		v := &fakeVerifier{enabled: false, ok: false, err: errors.New("must not be called")}
		a := &fakeAuthService{}
		w := post(newRouter(v, a), "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
		}
		if len(v.tokens) != 0 || a.loginCalls != 1 {
			t.Errorf("verify=%d login=%d, want 0 and 1", len(v.tokens), a.loginCalls)
		}
	})
}
