package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toRarityResponse(r *models.RarityResponse) gen.RarityResponse {
	if r == nil {
		return gen.RarityResponse{}
	}
	return gen.RarityResponse{
		Id:        uint32Ptr(uint32(r.ID)),
		Code:      strPtr(r.Code),
		Name:      strPtr(r.Name),
		CreatedAt: timePtr(r.CreatedAt),
		UpdatedAt: timePtr(r.UpdatedAt),
	}
}

func fromCreateRarityRequest(r *gen.CreateRarityRequest) *models.CreateRarityRequest {
	if r == nil {
		return nil
	}
	return &models.CreateRarityRequest{
		Code: r.Code,
		Name: r.Name,
	}
}

func fromUpdateRarityRequest(r *gen.UpdateRarityRequest) *models.UpdateRarityRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateRarityRequest{
		Code: r.Code,
		Name: r.Name,
	}
}
