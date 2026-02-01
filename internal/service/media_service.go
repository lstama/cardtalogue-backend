package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
	"github.com/lstama/cardtalogue-backend/internal/utils"
)

type MediaService interface {
	UploadImage(file *multipart.FileHeader, baseURL, uploadPath string, alsoScan bool) (*models.MediaResponse, map[string]interface{}, error)
	GetMedia(id uint) (*models.Media, error)
	DeleteMedia(id uint) error
	DeletePendingMedia() error
	DeleteOldPendingMedia(minutes int) error
	// ScanImage(file *multipart.FileHeader) (*models.CardScanResponse, error)
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
	if ValidateImage(file, 5*1024*1024) != nil {
		return nil, nil, errors.New("invalid image file")
	}

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

// func (s *mediaService) ScanImage(file *multipart.FileHeader) (*models.CardScanResponse, error) {
// 	if ValidateImage(file, 5*1024*1024) != nil {
// 		return nil, errors.New("invalid image file")
// 	}

// 	var response *models.CardScanResponse
// 	if s.recognitionSvc != nil {
// 		scanData, err := s.recognitionSvc.RecognizeCard(file)
// 		if err != nil {
// 			return nil, fmt.Errorf("card recognition failed: %w", err)
// 		}
// 		response = scanData
// 	}
// 	return response, nil
// }

func (s *mediaService) GetMedia(id uint) (*models.Media, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *mediaService) DeleteMedia(id uint) error {
	return s.mediaRepo.Delete(id)
}

func (s *mediaService) DeletePendingMedia() error {
	return s.mediaRepo.DeleteByStatus("pending")
}

func (s *mediaService) DeleteOldPendingMedia(minutes int) error {
	return s.mediaRepo.DeleteByStatusOlderThan("pending", minutes)
}

func ValidateImage(header *multipart.FileHeader, maxBytes int64) error {
	// 1. Check Size
	if header.Size > maxBytes {
		return errors.New("file size exceeds limit")
	}

	// 2. Open file to sniff content
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the first 512 bytes to detect the type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return err
	}

	fileType := http.DetectContentType(buff)

	// 3. Allowed Mime Types
	switch fileType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return errors.New("invalid file type: " + fileType)
	}
}
