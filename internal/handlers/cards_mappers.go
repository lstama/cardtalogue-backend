package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

// ---- Conversions: models -> contract DTOs ----

func toCardResponse(c *models.CardResponse) gen.CardResponse {
	if c == nil {
		return gen.CardResponse{}
	}
	return gen.CardResponse{
		Id:              uint32Ptr(uint32(c.ID)),
		Code:            strPtr(c.Code),
		Name:            strPtr(c.Name),
		ImageUrl:        strPtr(c.ImageURL),
		Description:     strPtr(c.Description),
		Status:          cardStatusPtr(c.Status),
		ForSale:         cardForSalePtr(c.ForSale),
		GameId:          uint32Ptr(uint32(c.GameID)),
		GameName:        strPtr(c.GameName),
		LanguageId:      uint32Ptr(uint32(c.LanguageID)),
		LanguageName:    strPtr(c.LanguageName),
		LanguageCode:    strPtr(c.LanguageCode),
		PackId:          uint32Ptr(uint32(c.PackID)),
		PackName:        strPtr(c.PackName),
		PackCode:        strPtr(c.PackCode),
		RarityId:        uint32Ptr(uint32(c.RarityID)),
		RarityName:      strPtr(c.RarityName),
		RarityCode:      strPtr(c.RarityCode),
		MediaId:         uint32PtrFromUint(c.MediaID),
		Price:           uint32Ptr(uint32(c.Price)),
		PriceMultiplier: uint32Ptr(uint32(c.PriceMultiplier)),
		Quantity:        uint32Ptr(uint32(c.Quantity)),
		CreatedAt:       timePtr(c.CreatedAt),
		UpdatedAt:       timePtr(c.UpdatedAt),
	}
}

func toCardProtectedResponse(c *models.CardProtectedResponse) gen.CardProtectedResponse {
	if c == nil {
		return gen.CardProtectedResponse{}
	}
	return gen.CardProtectedResponse{
		Id:              uint32Ptr(uint32(c.ID)),
		Code:            strPtr(c.Code),
		Name:            strPtr(c.Name),
		ImageUrl:        strPtr(c.ImageURL),
		Description:     strPtr(c.Description),
		Status:          cardProtectedStatusPtr(c.Status),
		ForSale:         cardProtectedForSalePtr(c.ForSale),
		GameId:          uint32Ptr(uint32(c.GameID)),
		GameName:        strPtr(c.GameName),
		LanguageId:      uint32Ptr(uint32(c.LanguageID)),
		LanguageName:    strPtr(c.LanguageName),
		LanguageCode:    strPtr(c.LanguageCode),
		PackId:          uint32Ptr(uint32(c.PackID)),
		PackName:        strPtr(c.PackName),
		PackCode:        strPtr(c.PackCode),
		RarityId:        uint32Ptr(uint32(c.RarityID)),
		RarityName:      strPtr(c.RarityName),
		RarityCode:      strPtr(c.RarityCode),
		StorageId:       uint32Ptr(uint32(c.StorageID)),
		StorageName:     strPtr(c.StorageName),
		StorageCode:     strPtr(c.StorageCode),
		MediaId:         uint32PtrFromUint(c.MediaID),
		Price:           uint32Ptr(uint32(c.Price)),
		PriceMultiplier: uint32Ptr(uint32(c.PriceMultiplier)),
		Quantity:        uint32Ptr(uint32(c.Quantity)),
		CreatedAt:       timePtr(c.CreatedAt),
		UpdatedAt:       timePtr(c.UpdatedAt),
	}
}

// toCardProtectedResponseFromCard converts a non-protected CardResponse to the
// protected contract shape, leaving storage fields unset.

func toCardProtectedResponseFromCard(c *models.CardResponse) gen.CardProtectedResponse {
	if c == nil {
		return gen.CardProtectedResponse{}
	}
	return gen.CardProtectedResponse{
		Id:              uint32Ptr(uint32(c.ID)),
		Code:            strPtr(c.Code),
		Name:            strPtr(c.Name),
		ImageUrl:        strPtr(c.ImageURL),
		Description:     strPtr(c.Description),
		Status:          cardProtectedStatusPtr(c.Status),
		ForSale:         cardProtectedForSalePtr(c.ForSale),
		GameId:          uint32Ptr(uint32(c.GameID)),
		GameName:        strPtr(c.GameName),
		LanguageId:      uint32Ptr(uint32(c.LanguageID)),
		LanguageName:    strPtr(c.LanguageName),
		LanguageCode:    strPtr(c.LanguageCode),
		PackId:          uint32Ptr(uint32(c.PackID)),
		PackName:        strPtr(c.PackName),
		PackCode:        strPtr(c.PackCode),
		RarityId:        uint32Ptr(uint32(c.RarityID)),
		RarityName:      strPtr(c.RarityName),
		RarityCode:      strPtr(c.RarityCode),
		MediaId:         uint32PtrFromUint(c.MediaID),
		Price:           uint32Ptr(uint32(c.Price)),
		PriceMultiplier: uint32Ptr(uint32(c.PriceMultiplier)),
		Quantity:        uint32Ptr(uint32(c.Quantity)),
		CreatedAt:       timePtr(c.CreatedAt),
		UpdatedAt:       timePtr(c.UpdatedAt),
	}
}

// ---- Conversions: contract requests -> models ----

func fromCreateCardRequest(r *gen.CreateCardRequest) *models.CreateCardRequest {
	if r == nil {
		return nil
	}
	return &models.CreateCardRequest{
		Code:            r.Code,
		Description:     derefStr(r.Description),
		ForSale:         derefCardForSale(r.ForSale),
		GameID:          uint(r.GameId),
		ImageURL:        derefStr(r.ImageUrl),
		LanguageID:      derefUint32(r.LanguageId),
		MediaID:         uintPtrFromUint32(r.MediaId),
		Name:            r.Name,
		PackID:          derefUint32(r.PackId),
		Price:           derefUint32(r.Price),
		PriceMultiplier: derefUint32(r.PriceMultiplier),
		Quantity:        derefUint32(r.Quantity),
		RarityID:        derefUint32(r.RarityId),
		Status:          derefCardStatus(r.Status),
		StorageID:       derefUint32(r.StorageId),
	}
}

func fromUpdateCardRequest(r *gen.UpdateCardRequest) *models.UpdateCardRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateCardRequest{
		Code:            r.Code,
		Description:     r.Description,
		ForSale:         derefCardForSalePtr(r.ForSale),
		GameID:          uintPtrFromUint32(r.GameId),
		ImageURL:        r.ImageUrl,
		LanguageID:      uintPtrFromUint32(r.LanguageId),
		MediaID:         uintPtrFromUint32(r.MediaId),
		Name:            r.Name,
		PackID:          uintPtrFromUint32(r.PackId),
		Price:           uintPtrFromUint32(r.Price),
		PriceMultiplier: uintPtrFromUint32(r.PriceMultiplier),
		Quantity:        uintPtrFromUint32(r.Quantity),
		RarityID:        uintPtrFromUint32(r.RarityId),
		Status:          derefCardStatusPtr(r.Status),
		StorageID:       uintPtrFromUint32(r.StorageId),
	}
}
