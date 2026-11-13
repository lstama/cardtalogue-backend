package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toGameResponse(g *models.GameResponse) gen.GameResponse {
	if g == nil {
		return gen.GameResponse{}
	}
	return gen.GameResponse{
		Id:        uint32Ptr(uint32(g.ID)),
		Name:      strPtr(g.Name),
		CreatedAt: timePtr(g.CreatedAt),
		UpdatedAt: timePtr(g.UpdatedAt),
	}
}

func fromCreateGameRequest(r *gen.CreateGameRequest) *models.CreateGameRequest {
	if r == nil {
		return nil
	}
	return &models.CreateGameRequest{
		Name: r.Name,
	}
}

func fromUpdateGameRequest(r *gen.UpdateGameRequest) *models.UpdateGameRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateGameRequest{
		Name: r.Name,
	}
}
