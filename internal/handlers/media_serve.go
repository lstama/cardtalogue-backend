package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// UploadImage godoc
// @Summary Upload an image
// @Security BearerAuth
// @Description Upload an image file. Image will be converted to JPEG format.
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file (JPEG, PNG, GIF)"
// @Param also_scan formData boolean false "Whether to scan the image (default: false)"
// @Success 201 {object} object{data=models.MediaResponse,scan=object}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /media/upload [post]
func (s *shared) UploadImage(c *gin.Context) {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Get also_scan parameter (optional)
	alsoScan := c.DefaultQuery("also_scan", "false") == "true"

	// Upload image
	media, scanResult, err := s.services.Media.UploadImage(file, s.cfg.APIBaseURL, s.cfg.MediaUploadPath, alsoScan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := gin.H{"data": media}
	if scanResult != nil {
		response["scan"] = scanResult
	}

	c.JSON(http.StatusCreated, response)
}

// ScanImage godoc
// @Summary Scan an image for card recognition
// @Description Scan an uploaded image file to recognize card details
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file (JPEG, PNG, GIF)"
// @Param X-Cap-Token header string true "Cap CAPTCHA token"
// @Success 200 {object} object{data=models.CardScanResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /media/scan [post]
func (s *shared) ScanImage(c *gin.Context) {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Scan image
	scanResult, err := s.services.Media.ScanImage(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scanResult})
}

// ServeMedia godoc
// @Summary Serve an image file
// @Description Retrieve and serve a media file by filename
// @Tags Media
// @Produce image/jpeg
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /media/{filename} [get]
func (s *shared) ServeMedia(c *gin.Context) {
	filename := c.Param("filename")

	// Validate filename to prevent path traversal
	if !isValidFilename(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	filePath := filepath.Join(s.cfg.MediaUploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media file not found"})
		return
	}

	c.File(filePath)
}

// isValidFilename checks if a filename is valid (prevents path traversal)
func isValidFilename(filename string) bool {
	// Check for path traversal attempts
	if filename == "" || filename == "." || filename == ".." {
		return false
	}
	if filepath.IsAbs(filename) {
		return false
	}
	if filepath.Clean(filename) != filename {
		return false
	}
	return true
}
