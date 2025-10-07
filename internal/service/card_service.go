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
	GetCardProtected(id uint) (*models.CardProtectedResponse, error)
	GetCards(limit, offset int) ([]*models.CardResponse, int64, error)
	GetCardsProtected(limit, offset int) ([]*models.CardProtectedResponse, int64, error)
	UpdateCard(id uint, req *models.UpdateCardRequest) (*models.CardResponse, error)
	DeleteCard(id uint) error
	SearchCards(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error)
}

type cardService struct {
	cardRepo repository.CardRepository
}

func NewCardService(cardRepo repository.CardRepository) CardService {
	return &cardService{cardRepo: cardRepo}
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
	card := &models.Card{
		Code:            req.Code,
		Name:            req.Name,
		Price:           req.Price,
		PriceMultiplier: req.PriceMultiplier,
		Quantity:        req.Quantity,
		ImageURL:        req.ImageURL,
		Description:     req.Description,
		Status:          status,
		ForSale:         forSale,
		GameID:          req.GameID,
		LanguageID:      req.LanguageID,
		PackID:          req.PackID,
		RarityID:        req.RarityID,
		StorageID:       req.StorageID,
		LowercaseCode:   utils.Normalize(req.Code),
		LowercaseName:   utils.Normalize(req.Name),
	}

	if err := s.cardRepo.Create(card); err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
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
		limit = 20
	}
	if limit > 100 {
		limit = 100
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

func (s *cardService) GetCardsProtected(limit, offset int) ([]*models.CardProtectedResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
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
	if req.ImageURL != nil {
		card.ImageURL = *req.ImageURL
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
		CreatedAt:       card.CreatedAt,
		UpdatedAt:       card.UpdatedAt,
	}
}

func (s *cardService) SearchCards(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.CardResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
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
