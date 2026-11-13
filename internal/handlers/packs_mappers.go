package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toPackResponse(p *models.PackResponse) gen.PackResponse {
	if p == nil {
		return gen.PackResponse{}
	}
	return gen.PackResponse{
		Id:           uint32Ptr(uint32(p.ID)),
		Code:         strPtr(p.Code),
		Name:         strPtr(p.Name),
		Description:  strPtr(p.Description),
		ImageUrl:     strPtr(p.ImageURL),
		GameId:       uint32Ptr(uint32(p.GameID)),
		GameName:     strPtr(p.GameName),
		LanguageId:   uint32Ptr(uint32(p.LanguageID)),
		LanguageName: strPtr(p.LanguageName),
		LanguageCode: strPtr(p.LanguageCode),
		CreatedAt:    timePtr(p.CreatedAt),
		UpdatedAt:    timePtr(p.UpdatedAt),
	}
}

func fromCreatePackRequest(r *gen.CreatePackRequest) *models.CreatePackRequest {
	if r == nil {
		return nil
	}
	return &models.CreatePackRequest{
		Code:        r.Code,
		Name:        r.Name,
		ImageURL:    r.ImageUrl,
		Description: derefStr(r.Description),
		GameID:      uint(r.GameId),
		LanguageID:  derefUint32(r.LanguageId),
	}
}

func fromUpdatePackRequest(r *gen.UpdatePackRequest) *models.UpdatePackRequest {
	if r == nil {
		return nil
	}
	return &models.UpdatePackRequest{
		Code:       r.Code,
		Name:       r.Name,
		ImageURL:   r.ImageUrl,
		Status:     derefUpdatePackStatusPtr(r.Status),
		GameID:     uintPtrFromUint32(r.GameId),
		LanguageID: uintPtrFromUint32(r.LanguageId),
	}
}
