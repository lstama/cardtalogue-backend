package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetRarities(ctx context.Context, req gen.GetRaritiesRequestObject) (gen.GetRaritiesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	rarities, _, err := s.services.Rarities.GetRarities(limit, offset)
	if err != nil {
		return gen.GetRarities500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.RarityResponse, 0, len(rarities))
	for _, r := range rarities {
		out = append(out, toRarityResponse(r))
	}
	return gen.GetRarities200JSONResponse(out), nil
}

func (s *shared) SearchRarities(ctx context.Context, req gen.SearchRaritiesRequestObject) (gen.SearchRaritiesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	rarities, _, err := s.services.Rarities.SearchRarities(derefStr(req.Params.Query), limit, offset)
	if err != nil {
		return gen.SearchRarities500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.RarityResponse, 0, len(rarities))
	for _, r := range rarities {
		out = append(out, toRarityResponse(r))
	}
	return gen.SearchRarities200JSONResponse(out), nil
}

func (s *shared) GetRarityById(ctx context.Context, req gen.GetRarityByIdRequestObject) (gen.GetRarityByIdResponseObject, error) {
	r, err := s.services.Rarities.GetRarity(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetRarityById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetRarityById500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetRarityById200JSONResponse(toRarityResponse(r)), nil
}

func (s *shared) CreateRarity(ctx context.Context, req gen.CreateRarityRequestObject) (gen.CreateRarityResponseObject, error) {
	r, err := s.services.Rarities.CreateRarity(fromCreateRarityRequest(req.Body))
	if err != nil {
		return gen.CreateRarity400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateRarity200JSONResponse(toRarityResponse(r)), nil
}

func (s *shared) UpdateRarity(ctx context.Context, req gen.UpdateRarityRequestObject) (gen.UpdateRarityResponseObject, error) {
	r, err := s.services.Rarities.UpdateRarity(uint(req.Id), fromUpdateRarityRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateRarity404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateRarity400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateRarity200JSONResponse(toRarityResponse(r)), nil
}

func (s *shared) DeleteRarity(ctx context.Context, req gen.DeleteRarityRequestObject) (gen.DeleteRarityResponseObject, error) {
	if err := s.services.Rarities.DeleteRarity(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteRarity404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteRarity404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteRarity200Response{}, nil
}
