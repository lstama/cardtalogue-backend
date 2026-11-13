package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
	"github.com/lstama/cardtalogue-backend/internal/utils"
)

type CardService interface {
	CreateCard(req *models.CreateCardRequest) (*models.CardResponse, error)
	GetCard(id uint) (*models.CardResponse, error)
	GetCardPublic(id uint) (*models.CardResponse, error)
	GetCardProtected(id uint) (*models.CardProtectedResponse, error)
	GetCards(limit, offset int) ([]*models.CardResponse, int64, error)
	GetCardsPublic(limit, offset int) ([]*models.CardResponse, int64, error)
	GetCardsProtected(limit, offset int) ([]*models.CardProtectedResponse, int64, error)
	UpdateCard(id uint, req *models.UpdateCardRequest) (*models.CardResponse, error)
	DeleteCard(id uint) error
	SearchCards(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error)
	SearchCardsPublic(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error)
}

type cardService struct {
	cardRepo  repository.CardRepository
	mediaRepo repository.MediaRepository
}

func NewCardService(cardRepo repository.CardRepository, mediaRepo repository.MediaRepository) CardService {
	return &cardService{cardRepo: cardRepo, mediaRepo: mediaRepo}
}

func (s *cardService) CreateCard(req *models.CreateCardRequest) (*models.CardResponse, error) {
	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = "enabled"
	}
	forSale := req.ForSale
	if forSale == "" {
		forSale = "yes"
	}

	// Handle media/imageURL
	var mediaID *uint
	var imageURL string

	if req.MediaID != nil {
		// If mediaID is provided, get the media and use its URL
		media, err := s.mediaRepo.GetByID(*req.MediaID)
		if err != nil {
			return nil, fmt.Errorf("failed to get media: %w", err)
		}
		imageURL = media.URL
		mediaID = req.MediaID

		// Mark media as attached
		media.Status = "attached"
		if err := s.mediaRepo.Update(media); err != nil {
			return nil, fmt.Errorf("failed to update media status: %w", err)
		}
	} else if req.ImageURL != "" {
		// If imageURL is provided without mediaID, create a new media entry
		newMedia := &models.Media{
			URL:      req.ImageURL,
			Status:   "attached",
			Provider: "hardlink",
		}
		if err := s.mediaRepo.Create(newMedia); err != nil {
			return nil, fmt.Errorf("failed to create media: %w", err)
		}
		imageURL = newMedia.URL
		mediaID = &newMedia.ID
	} else {
		// If no mediaID and no imageURL provided, use default with hardlink provider
		defaultURL := "https://picsum.photos/230/330.webp"
		newMedia := &models.Media{
			URL:      defaultURL,
			Status:   "attached",
			Provider: "hardlink",
		}
		if err := s.mediaRepo.Create(newMedia); err != nil {
			return nil, fmt.Errorf("failed to create media: %w", err)
		}
		imageURL = newMedia.URL
		mediaID = &newMedia.ID
	}

	card := &models.Card{
		Code:            req.Code,
		Name:            req.Name,
		Price:           req.Price,
		PriceMultiplier: req.PriceMultiplier,
		Quantity:        req.Quantity,
		ImageURL:        imageURL,
		Description:     req.Description,
		Status:          status,
		ForSale:         forSale,
		GameID:          req.GameID,
		LanguageID:      req.LanguageID,
		PackID:          req.PackID,
		RarityID:        req.RarityID,
		StorageID:       req.StorageID,
		MediaID:         mediaID,
		LowercaseCode:   utils.Normalize(req.Code),
		LowercaseName:   utils.Normalize(req.Name),
	}

	if err := s.cardRepo.Create(card); err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	// Delete pending media older than 10 minutes after successfully creating the card
	// This avoids deleting photos that other users are currently working on
	if err := s.mediaRepo.DeleteByStatusOlderThan("pending", 10); err != nil {
		// Log the error but don't fail the card creation
		fmt.Printf("Warning: failed to delete old pending media: %v\n", err)
	}

	return s.toCardResponse(card), nil
}

func (s *cardService) GetCard(id uint) (*models.CardResponse, error) {
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toCardResponse(card), nil
}

func (s *cardService) GetCardPublic(id uint) (*models.CardResponse, error) {
	card, err := s.cardRepo.GetByIDWithStatus(id, false)
	if err != nil {
		return nil, err
	}
	return s.toCardResponse(card), nil
}

func (s *cardService) GetCardProtected(id uint) (*models.CardProtectedResponse, error) {
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toCardProtectedResponse(card), nil
}

func (s *cardService) GetCards(limit, offset int) ([]*models.CardResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cards, err := s.cardRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.cardRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CardResponse, len(cards))
	for i, card := range cards {
		responses[i] = s.toCardResponse(card)
	}

	return responses, total, nil
}

func (s *cardService) GetCardsPublic(limit, offset int) ([]*models.CardResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cards, err := s.cardRepo.GetAllWithStatus(limit, offset, false)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.cardRepo.CountWithStatus(false)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CardResponse, len(cards))
	for i, card := range cards {
		responses[i] = s.toCardResponse(card)
	}

	return responses, total, nil
}

func (s *cardService) GetCardsProtected(limit, offset int) ([]*models.CardProtectedResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cards, err := s.cardRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.cardRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CardProtectedResponse, len(cards))
	for i, card := range cards {
		responses[i] = s.toCardProtectedResponse(card)
	}

	return responses, total, nil
}

func (s *cardService) UpdateCard(id uint, req *models.UpdateCardRequest) (*models.CardResponse, error) {
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Code != nil {
		card.Code = *req.Code
		card.LowercaseCode = utils.Normalize(card.Code)
	}
	if req.Name != nil {
		card.Name = *req.Name
		card.LowercaseName = utils.Normalize(card.Name)
	}
	if req.Price != nil {
		card.Price = *req.Price
	}
	if req.PriceMultiplier != nil {
		card.PriceMultiplier = *req.PriceMultiplier
	}
	if req.Quantity != nil {
		card.Quantity = *req.Quantity
	}
	if req.Description != nil {
		card.Description = *req.Description
	}
	if req.Status != nil {
		card.Status = *req.Status
	}
	if req.ForSale != nil {
		card.ForSale = *req.ForSale
	}
	if req.GameID != nil {
		card.GameID = *req.GameID
	}
	if req.LanguageID != nil {
		card.LanguageID = *req.LanguageID
	}
	if req.PackID != nil {
		card.PackID = *req.PackID
	}
	if req.RarityID != nil {
		card.RarityID = *req.RarityID
	}
	if req.StorageID != nil {
		card.StorageID = *req.StorageID
	}

	// Handle media update
	if req.MediaID != nil {
		// If mediaID is being changed
		if card.MediaID == nil || *card.MediaID != *req.MediaID {
			// Archive the old media if it existed
			if card.MediaID != nil {
				oldMedia, err := s.mediaRepo.GetByID(*card.MediaID)
				if err == nil {
					oldMedia.Status = "archived"
					s.mediaRepo.Update(oldMedia)
				}
			}

			// Get the new media and mark as attached
			newMedia, err := s.mediaRepo.GetByID(*req.MediaID)
			if err != nil {
				return nil, fmt.Errorf("failed to get media: %w", err)
			}

			newMedia.Status = "attached"
			if err := s.mediaRepo.Update(newMedia); err != nil {
				return nil, fmt.Errorf("failed to update media status: %w", err)
			}

			card.MediaID = req.MediaID
			card.ImageURL = newMedia.URL
		}
	} else if req.ImageURL != nil {
		// Only update media if imageURL is different from current attached media's URL
		shouldUpdateMedia := true

		// Check if the imageURL is the same as the current attached media's URL
		if card.MediaID != nil {
			currentMedia, err := s.mediaRepo.GetByID(*card.MediaID)
			if err == nil && currentMedia.URL == *req.ImageURL {
				// URL is the same, don't create new media
				shouldUpdateMedia = false
			}
		}

		if shouldUpdateMedia {
			// Archive the old media if it existed
			if card.MediaID != nil {
				oldMedia, err := s.mediaRepo.GetByID(*card.MediaID)
				if err == nil {
					oldMedia.Status = "archived"
					s.mediaRepo.Update(oldMedia)
				}
			}

			// Create a new media entry with the imageURL
			newMedia := &models.Media{
				URL:      *req.ImageURL,
				Status:   "attached",
				Provider: "hardlink",
			}
			if err := s.mediaRepo.Create(newMedia); err != nil {
				return nil, fmt.Errorf("failed to create media: %w", err)
			}

			card.MediaID = &newMedia.ID
			card.ImageURL = newMedia.URL
		}
	}

	if err := s.cardRepo.Update(card); err != nil {
		return nil, fmt.Errorf("failed to update card: %w", err)
	}

	return s.toCardResponse(card), nil
}

func (s *cardService) DeleteCard(id uint) error {
	// Check if card exists
	if _, err := s.cardRepo.GetByID(id); err != nil {
		return err
	}

	return s.cardRepo.Delete(id)
}

func (s *cardService) toCardResponse(card *models.Card) *models.CardResponse {
	return &models.CardResponse{
		ID:              card.ID,
		Code:            card.Code,
		Name:            card.Name,
		Price:           card.Price,
		PriceMultiplier: card.PriceMultiplier,
		Quantity:        card.Quantity,
		ImageURL:        card.ImageURL,
		Description:     card.Description,
		Status:          card.Status,
		ForSale:         card.ForSale,
		GameID:          card.GameID,
		GameName:        card.Game.Name,
		LanguageID:      card.LanguageID,
		LanguageName:    card.Language.Name,
		LanguageCode:    card.Language.Code,
		PackID:          card.Pack.ID,
		PackName:        card.Pack.Name,
		PackCode:        card.Pack.Code,
		RarityID:        card.RarityID,
		RarityName:      card.Rarity.Name,
		RarityCode:      card.Rarity.Code,
		MediaID:         card.MediaID,
		CreatedAt:       card.CreatedAt,
		UpdatedAt:       card.UpdatedAt,
	}
}

// With sensitive info (storage details)
func (s *cardService) toCardProtectedResponse(card *models.Card) *models.CardProtectedResponse {
	return &models.CardProtectedResponse{
		ID:              card.ID,
		Code:            card.Code,
		Name:            card.Name,
		Price:           card.Price,
		PriceMultiplier: card.PriceMultiplier,
		Quantity:        card.Quantity,
		ImageURL:        card.ImageURL,
		Description:     card.Description,
		Status:          card.Status,
		ForSale:         card.ForSale,
		GameID:          card.GameID,
		GameName:        card.Game.Name,
		LanguageID:      card.LanguageID,
		LanguageName:    card.Language.Name,
		LanguageCode:    card.Language.Code,
		PackID:          card.Pack.ID,
		PackName:        card.Pack.Name,
		PackCode:        card.Pack.Code,
		RarityID:        card.RarityID,
		RarityName:      card.Rarity.Name,
		RarityCode:      card.Rarity.Code,
		StorageID:       card.StorageID,
		StorageName:     card.Storage.Name,
		StorageCode:     card.Storage.Code,
		MediaID:         card.MediaID,
		CreatedAt:       card.CreatedAt,
		UpdatedAt:       card.UpdatedAt,
	}
}

func (s *cardService) SearchCards(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cards, total, err := s.cardRepo.Search(game_id, language_id, pack_id, query_string, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CardResponse, len(cards))
	for i, card := range cards {
		responses[i] = s.toCardResponse(card)
	}

	return responses, total, nil
}

func (s *cardService) SearchCardsPublic(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cards, total, err := s.cardRepo.SearchWithStatus(game_id, language_id, pack_id, query_string, limit, offset, false)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CardResponse, len(cards))
	for i, card := range cards {
		responses[i] = s.toCardResponse(card)
	}

	return responses, total, nil
}
