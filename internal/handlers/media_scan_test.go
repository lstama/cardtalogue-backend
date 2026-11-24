package handlers

import (
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

// scanMultipartReader builds a one-shot multipart reader with a single file part,
// as the generated strict handler would from an uploaded request.
func scanMultipartReader(t *testing.T) *multipart.Reader {
	t.Helper()
	body, contentType := multipartForm(t, "file", "card.png", "fake-image-bytes")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("ParseMediaType(%q): %v", contentType, err)
	}
	return multipart.NewReader(body, params["boundary"])
}

func scanMediaRequest(token *string, body *multipart.Reader) gen.ScanMediaRequestObject {
	obj := gen.ScanMediaRequestObject{Body: body}
	if token != nil {
		capToken := gen.CapToken(*token)
		obj.Params.XCapToken = &capToken
	}
	return obj
}

func newScanShared(t *testing.T, v *fakeVerifier) (*shared, *int) {
	t.Helper()
	scanCalls := 0
	ms := &mockMediaService{
		scanFn: func(*multipart.FileHeader) (*models.CardScanResponse, error) {
			scanCalls++
			return &models.CardScanResponse{Success: true, ProcessingTime: 0.1}, nil
		},
	}
	return &shared{
		services: &Services{Media: ms},
		cfg:      &config.Config{},
		verifier: v,
	}, &scanCalls
}

// TestScanMediaUsesTypedCapToken proves the handler reads X-Cap-Token from the
// generated Params field rather than from a gin context type assertion: the
// request context here is a plain context.Context.
func TestScanMediaUsesTypedCapToken(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		token      *string
		verifyOK   bool
		verifyErr  error
		wantMsg    string
		wantStatus int
		wantVerify int
		wantScan   int
	}{
		{
			name:       "disabled verifier allows scan without token",
			enabled:    false,
			wantStatus: http.StatusOK,
			wantScan:   1,
		},
		{
			name:       "enabled verifier requires token",
			enabled:    true,
			wantMsg:    "captcha required",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "verifier error maps to unavailable",
			enabled:    true,
			token:      strPtr("cap.1.0.0:proof"),
			verifyErr:  errors.New("cap server unreachable"),
			wantMsg:    "captcha unavailable",
			wantStatus: http.StatusBadRequest,
			wantVerify: 1,
		},
		{
			name:       "rejected token maps to failed",
			enabled:    true,
			token:      strPtr("cap.1.0.0:proof"),
			verifyOK:   false,
			wantMsg:    "captcha failed",
			wantStatus: http.StatusBadRequest,
			wantVerify: 1,
		},
		{
			name:       "valid token proceeds to scan",
			enabled:    true,
			token:      strPtr("cap.1.0.0:proof"),
			verifyOK:   true,
			wantStatus: http.StatusOK,
			wantVerify: 1,
			wantScan:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &fakeVerifier{enabled: tc.enabled, ok: tc.verifyOK, err: tc.verifyErr}
			s, scanCalls := newScanShared(t, v)

			req := scanMediaRequest(tc.token, scanMultipartReader(t))
			resp, err := s.ScanMedia(context.Background(), req)
			if err != nil {
				t.Fatalf("ScanMedia: %v", err)
			}

			switch want := resp.(type) {
			case gen.ScanMedia400JSONResponse:
				if tc.wantMsg == "" {
					t.Fatalf("unexpected 400 %q", want.Error)
				}
				if want.Error != tc.wantMsg {
					t.Errorf("error = %q, want %q", want.Error, tc.wantMsg)
				}
			case gen.ScanMedia200JSONResponse:
				if tc.wantMsg != "" {
					t.Fatalf("got 200 response, want 400 %q", tc.wantMsg)
				}
			default:
				t.Fatalf("response type = %T", resp)
			}

			w := httptest.NewRecorder()
			if err := resp.VisitScanMediaResponse(w); err != nil {
				t.Fatalf("VisitScanMediaResponse: %v", err)
			}
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if len(v.tokens) != tc.wantVerify {
				t.Errorf("verifier.Verify called %d times, want %d", len(v.tokens), tc.wantVerify)
			}
			if len(v.tokens) > 0 && v.tokens[0] != *tc.token {
				t.Errorf("Verify token = %q, want %q", v.tokens[0], *tc.token)
			}
			if *scanCalls != tc.wantScan {
				t.Errorf("Media.ScanImage called %d times, want %d", *scanCalls, tc.wantScan)
			}
		})
	}
}
