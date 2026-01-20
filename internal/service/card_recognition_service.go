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
	RecognizeCard(file *multipart.FileHeader) (*models.ScanResult, error)
}

type cardRecognitionService struct {
	apiURL string
}

func NewCardRecognitionService(apiURL string) CardRecognitionService {
	return &cardRecognitionService{apiURL: apiURL}
}

func (s *cardRecognitionService) RecognizeCard(file *multipart.FileHeader) (*models.ScanResult, error) {
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

	// Convert to simplified result
	result := &models.ScanResult{
		Matches: make([]models.ScanMatchResult, 0),
	}

	for _, match := range apiResponse.Matches {
		matchResult := models.ScanMatchResult{
			Name:       match.Name,
			Confidence: match.Confidence,
		}

		// Extract description from the info
		if match.Info != nil {
			descData := map[string]interface{}{}
			if match.Info.Description.Effect != nil {
				descData["effect"] = match.Info.Description.Effect
			}
			if match.Info.Description.MonsterEffect != nil {
				descData["monster_effect"] = match.Info.Description.MonsterEffect
			}
			if match.Info.Description.PendulumEffect != nil {
				descData["pendulum_effect"] = match.Info.Description.PendulumEffect
			}

			if len(descData) > 0 {
				descBytes, _ := json.Marshal(descData)
				matchResult.Description = descBytes
			}
		}

		result.Matches = append(result.Matches, matchResult)
	}

	return result, nil
}
