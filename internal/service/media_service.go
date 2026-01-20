package service

import (
	"fmt"
	"mime/multipart"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
	"github.com/lstama/cardtalogue-backend/internal/utils"
)

type MediaService interface {
	UploadImage(file *multipart.FileHeader, baseURL, uploadPath string, alsoScan bool) (*models.MediaResponse, map[string]interface{}, error)
	GetMedia(id uint) (*models.Media, error)
	DeleteMedia(id uint) error
}

type mediaService struct {
	mediaRepo      repository.MediaRepository
	recognitionSvc CardRecognitionService
}

func NewMediaService(mediaRepo repository.MediaRepository) MediaService {
	return &mediaService{mediaRepo: mediaRepo}
}

func NewMediaServiceWithRecognition(mediaRepo repository.MediaRepository, recognitionSvc CardRecognitionService) MediaService {
	return &mediaService{
		mediaRepo:      mediaRepo,
		recognitionSvc: recognitionSvc,
	}
}

func (s *mediaService) UploadImage(file *multipart.FileHeader, baseURL, uploadPath string, alsoScan bool) (*models.MediaResponse, map[string]interface{}, error) {
	// Process and save the image
	filename, err := utils.ProcessAndSaveImage(file, uploadPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Construct the full URL
	imageURL := fmt.Sprintf("%s/media/%s", baseURL, filename)

	// Create media entry in database
	media := &models.Media{
		URL:      imageURL,
		Status:   "attached",
		Provider: "local",
	}

	if err := s.mediaRepo.Create(media); err != nil {
		// Delete the file if database operation fails
		utils.DeleteImage(uploadPath, filename)
		return nil, nil, fmt.Errorf("failed to create media record: %w", err)
	}

	// Convert to response
	response := &models.MediaResponse{
		ID:        media.ID,
		URL:       media.URL,
		Status:    media.Status,
		Provider:  media.Provider,
		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,
	}

	// Handle card recognition if requested
	var scanResult map[string]interface{}
	if alsoScan && s.recognitionSvc != nil {
		scanData, err := s.recognitionSvc.RecognizeCard(file)
		if err != nil {
			// Log error but don't fail the upload
			fmt.Printf("Warning: card recognition failed: %v\n", err)
		} else if scanData != nil {
			scanResult = map[string]interface{}{
				"scan": scanData,
			}
		}
	}

	return response, scanResult, nil
}

func (s *mediaService) GetMedia(id uint) (*models.Media, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *mediaService) DeleteMedia(id uint) error {
	return s.mediaRepo.Delete(id)
}
