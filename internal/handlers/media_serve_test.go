package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"
)

// mockMediaService embeds the interface so only the methods under test need
// stubbing; any other method call panics (nil interface), which is fine because
// the media serve handlers never touch them.
type mockMediaService struct {
	service.MediaService
	uploadFn func(*multipart.FileHeader, string, string, bool) (*models.MediaResponse, map[string]interface{}, error)
	scanFn   func(*multipart.FileHeader) (*models.CardScanResponse, error)
}

// UploadImage delegates to uploadFn so the handler-under-test can be exercised
// without a real media service.
func (m *mockMediaService) UploadImage(file *multipart.FileHeader, baseURL, uploadPath string, alsoScan bool) (*models.MediaResponse, map[string]interface{}, error) {
	return m.uploadFn(file, baseURL, uploadPath, alsoScan)
}

// ScanImage delegates to scanFn.
func (m *mockMediaService) ScanImage(file *multipart.FileHeader) (*models.CardScanResponse, error) {
	return m.scanFn(file)
}

func newTestShared(ms service.MediaService) *shared {
	return &shared{
		services: &Services{Media: ms},
		cfg:      &config.Config{MediaUploadPath: "media-uploads", APIBaseURL: "http://media.test"},
		verifier: nil, // media serve handlers do not use the verifier
	}
}

func newMediaRouter(s *shared) *gin.Engine {
	r := gin.New()
	r.GET("/media/:filename", s.ServeMedia)
	r.POST("/media/upload", s.UploadImage)
	r.POST("/media/scan", s.ScanImage)
	return r
}

// multipartForm builds a multipart/form-data body with a single file field.
func multipartForm(t *testing.T, fieldName, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return body, w.FormDataContentType()
}

func doRequest(t *testing.T, r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestServeMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	s := newTestShared(&mockMediaService{})

	// Valid file with an allowed extension.
	if err := os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("BYTES"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.cfg.MediaUploadPath = dir

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"serves valid file", "/media/photo.jpg", http.StatusOK, "BYTES"},
		{"missing file", "/media/missing.jpg", http.StatusNotFound, `{"error":"media file not found"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(t, newMediaRouter(s), httptest.NewRequest(http.MethodGet, tc.path, nil))
			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body=%s)", w.Code, tc.wantStatus, w.Body.String())
			}
			if w.Body.String() != tc.wantBody {
				t.Fatalf("body = %q, want %q", w.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestUploadImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestShared(&mockMediaService{
		uploadFn: func(*multipart.FileHeader, string, string, bool) (*models.MediaResponse, map[string]interface{}, error) {
			return &models.MediaResponse{ID: 7, FileName: "photo.jpg"}, map[string]interface{}{"score": 0.9}, nil
		},
	})

	t.Run("success returns media and scan", func(t *testing.T) {
		body, ct := multipartForm(t, "file", "photo.jpg", "image-bytes")
		req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
		req.Header.Set("Content-Type", ct)
		w := doRequest(t, newMediaRouter(s), req)
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", w.Code, http.StatusCreated, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"id":7`) || !strings.Contains(w.Body.String(), `"score":0.9`) {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("missing file field is 400", func(t *testing.T) {
		w := doRequest(t, newMediaRouter(s), httptest.NewRequest(http.MethodPost, "/media/upload", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("service error is 500", func(t *testing.T) {
		s2 := newTestShared(&mockMediaService{
			uploadFn: func(*multipart.FileHeader, string, string, bool) (*models.MediaResponse, map[string]interface{}, error) {
				return nil, nil, errBoom
			},
		})
		body, ct := multipartForm(t, "file", "photo.jpg", "image-bytes")
		req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
		req.Header.Set("Content-Type", ct)
		w := doRequest(t, newMediaRouter(s2), req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestScanImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTestShared(&mockMediaService{
		scanFn: func(*multipart.FileHeader) (*models.CardScanResponse, error) {
			return &models.CardScanResponse{Success: true, ProcessingTime: 1.2}, nil
		},
	})

	t.Run("success returns scan", func(t *testing.T) {
		body, ct := multipartForm(t, "file", "photo.jpg", "image-bytes")
		req := httptest.NewRequest(http.MethodPost, "/media/scan", body)
		req.Header.Set("Content-Type", ct)
		w := doRequest(t, newMediaRouter(s), req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"success":true`) {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("missing file field is 400", func(t *testing.T) {
		w := doRequest(t, newMediaRouter(s), httptest.NewRequest(http.MethodPost, "/media/scan", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("service error is 500", func(t *testing.T) {
		s2 := newTestShared(&mockMediaService{
			scanFn: func(*multipart.FileHeader) (*models.CardScanResponse, error) {
				return nil, errBoom
			},
		})
		body, ct := multipartForm(t, "file", "photo.jpg", "image-bytes")
		req := httptest.NewRequest(http.MethodPost, "/media/scan", body)
		req.Header.Set("Content-Type", ct)
		w := doRequest(t, newMediaRouter(s2), req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestIsValidFilename(t *testing.T) {
	// Note: isValidFilename guards against path traversal only (it does not
	// validate file extensions) -- that is the original contract and is
	// preserved verbatim.
	valid := []string{"photo.jpg", "a/b/pic.PNG", "doc.pdf", "plain"}
	invalid := []string{"", ".", "..", "/etc/passwd", "a/../b", "x/y/../z"}
	for _, name := range valid {
		if !isValidFilename(name) {
			t.Errorf("isValidFilename(%q) = false, want true", name)
		}
	}
	for _, name := range invalid {
		if isValidFilename(name) {
			t.Errorf("isValidFilename(%q) = true, want false", name)
		}
	}
}

var errBoom = &boomError{}

type boomError struct{}

func (*boomError) Error() string { return "boom" }
