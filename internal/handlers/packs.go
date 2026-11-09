package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetPacks(ctx context.Context, req gen.GetPacksRequestObject) (gen.GetPacksResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	packs, _, err := s.services.Packs.GetPacks(limit, offset)
	if err != nil {
		return gen.GetPacks500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.PackResponse, 0, len(packs))
	for _, p := range packs {
		out = append(out, toPackResponse(p))
	}
	return gen.GetPacks200JSONResponse(out), nil
}

func (s *shared) SearchPacks(ctx context.Context, req gen.SearchPacksRequestObject) (gen.SearchPacksResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	packs, _, err := s.services.Packs.SearchPacks(int(derefUint32(req.Params.GameId)), int(derefUint32(req.Params.LanguageId)), derefStr(req.Params.Query), int(limit), int(offset))
	if err != nil {
		return gen.SearchPacks500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.PackResponse, 0, len(packs))
	for _, p := range packs {
		out = append(out, toPackResponse(p))
	}
	return gen.SearchPacks200JSONResponse(out), nil
}

func (s *shared) GetPackById(ctx context.Context, req gen.GetPackByIdRequestObject) (gen.GetPackByIdResponseObject, error) {
	p, err := s.services.Packs.GetPack(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetPackById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetPackById500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetPackById200JSONResponse(toPackResponse(p)), nil
}

func (s *shared) CreatePack(ctx context.Context, req gen.CreatePackRequestObject) (gen.CreatePackResponseObject, error) {
	p, err := s.services.Packs.CreatePack(fromCreatePackRequest(req.Body))
	if err != nil {
		return gen.CreatePack400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreatePack200JSONResponse(toPackResponse(p)), nil
}

func (s *shared) UpdatePack(ctx context.Context, req gen.UpdatePackRequestObject) (gen.UpdatePackResponseObject, error) {
	p, err := s.services.Packs.UpdatePack(uint(req.Id), fromUpdatePackRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdatePack404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdatePack400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdatePack200JSONResponse(toPackResponse(p)), nil
}

func (s *shared) DeletePack(ctx context.Context, req gen.DeletePackRequestObject) (gen.DeletePackResponseObject, error) {
	if err := s.services.Packs.DeletePack(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeletePack404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeletePack404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeletePack200Response{}, nil
}
