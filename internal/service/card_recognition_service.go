package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/lstama/cardtalogue-backend/internal/models"
)

type CardRecognitionService interface {
	RecognizeCard(file *multipart.FileHeader) (*models.CardScanResponse, error)
}

type cardRecognitionService struct {
	apiURL string
}

func NewCardRecognitionService(apiURL string) CardRecognitionService {
	return &cardRecognitionService{apiURL: apiURL}
}

func (s *cardRecognitionService) RecognizeCard(file *multipart.FileHeader) (*models.CardScanResponse, error) {
	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image file
	filePart, err := writer.CreateFormFile("image", file.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	if _, err := io.Copy(filePart, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Make request
	req, err := http.NewRequest("POST", s.apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("recognition API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse models.CardScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("recognition API returned success=false")
	}

	return &apiResponse, nil
}
