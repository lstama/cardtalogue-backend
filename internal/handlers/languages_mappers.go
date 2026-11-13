package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toLanguageResponse(l *models.LanguageResponse) gen.LanguageResponse {
	if l == nil {
		return gen.LanguageResponse{}
	}
	return gen.LanguageResponse{
		Id:        uint32Ptr(uint32(l.ID)),
		Code:      strPtr(l.Code),
		Name:      strPtr(l.Name),
		CreatedAt: timePtr(l.CreatedAt),
		UpdatedAt: timePtr(l.UpdatedAt),
	}
}

func fromCreateLanguageRequest(r *gen.CreateLanguageRequest) *models.CreateLanguageRequest {
	if r == nil {
		return nil
	}
	return &models.CreateLanguageRequest{
		Code: r.Code,
		Name: r.Name,
	}
}

func fromUpdateLanguageRequest(r *gen.UpdateLanguageRequest) *models.UpdateLanguageRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateLanguageRequest{
		Code: r.Code,
		Name: r.Name,
	}
}
