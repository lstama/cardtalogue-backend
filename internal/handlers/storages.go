package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetStorages(ctx context.Context, req gen.GetStoragesRequestObject) (gen.GetStoragesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	storages, _, err := s.services.Storages.GetStorages(limit, offset)
	if err != nil {
		return gen.GetStorages500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.StorageResponse, 0, len(storages))
	for _, st := range storages {
		out = append(out, toStorageResponse(st))
	}
	return gen.GetStorages200JSONResponse(out), nil
}

func (s *shared) SearchStorages(ctx context.Context, req gen.SearchStoragesRequestObject) (gen.SearchStoragesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	storages, _, err := s.services.Storages.SearchStorages(derefStr(req.Params.Query), limit, offset)
	if err != nil {
		return gen.SearchStorages500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.StorageResponse, 0, len(storages))
	for _, st := range storages {
		out = append(out, toStorageResponse(st))
	}
	return gen.SearchStorages200JSONResponse(out), nil
}

func (s *shared) GetStorageById(ctx context.Context, req gen.GetStorageByIdRequestObject) (gen.GetStorageByIdResponseObject, error) {
	st, err := s.services.Storages.GetStorage(uint(req.Id))
	if err != nil {
		return gen.GetStorageById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetStorageById200JSONResponse(toStorageResponse(st)), nil
}

func (s *shared) CreateStorage(ctx context.Context, req gen.CreateStorageRequestObject) (gen.CreateStorageResponseObject, error) {
	st, err := s.services.Storages.CreateStorage(fromCreateStorageRequest(req.Body))
	if err != nil {
		return gen.CreateStorage400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateStorage200JSONResponse(toStorageResponse(st)), nil
}

func (s *shared) UpdateStorage(ctx context.Context, req gen.UpdateStorageRequestObject) (gen.UpdateStorageResponseObject, error) {
	st, err := s.services.Storages.UpdateStorage(uint(req.Id), fromUpdateStorageRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateStorage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateStorage400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateStorage200JSONResponse(toStorageResponse(st)), nil
}

func (s *shared) DeleteStorage(ctx context.Context, req gen.DeleteStorageRequestObject) (gen.DeleteStorageResponseObject, error) {
	if err := s.services.Storages.DeleteStorage(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteStorage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteStorage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteStorage200Response{}, nil
}
