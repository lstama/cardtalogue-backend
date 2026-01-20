package utils

import (
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ProcessAndSaveImage processes an uploaded image, converts it to JPEG, and saves it
// Returns the filename (without path) and error
func ProcessAndSaveImage(file *multipart.FileHeader, uploadPath string) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate filename: hash(content) + uuid
	fileHash, err := hashFile(src)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	// Reset reader position after hashing
	if _, err := src.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek in file: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.jpeg", fileHash[:16], uuid.New().String()[:8])
	filePath := filepath.Join(uploadPath, filename)

	// Decode the image using the standard library
	img, _, err := image.Decode(src)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Reset reader position
	if _, err := src.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek in file: %w", err)
	}

	// Create output file
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Encode as JPEG
	opts := &jpeg.Options{Quality: 85}
	if err := jpeg.Encode(out, img, opts); err != nil {
		return "", fmt.Errorf("failed to encode image as JPEG: %w", err)
	}

	return filename, nil
}

// hashFile returns the first 16 characters of SHA256 hash of file content
func hashFile(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// DeleteImage removes an image file from the upload directory
func DeleteImage(uploadPath, filename string) error {
	filePath := filepath.Join(uploadPath, filename)
	return os.Remove(filePath)
}
