package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetLanguages(ctx context.Context, req gen.GetLanguagesRequestObject) (gen.GetLanguagesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	languages, _, err := s.services.Languages.GetLanguages(limit, offset)
	if err != nil {
		return gen.GetLanguages500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.LanguageResponse, 0, len(languages))
	for _, l := range languages {
		out = append(out, toLanguageResponse(l))
	}
	return gen.GetLanguages200JSONResponse(out), nil
}

func (s *shared) SearchLanguages(ctx context.Context, req gen.SearchLanguagesRequestObject) (gen.SearchLanguagesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	languages, _, err := s.services.Languages.SearchLanguages(derefStr(req.Params.Query), limit, offset)
	if err != nil {
		return gen.SearchLanguages500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.LanguageResponse, 0, len(languages))
	for _, l := range languages {
		out = append(out, toLanguageResponse(l))
	}
	return gen.SearchLanguages200JSONResponse(out), nil
}

func (s *shared) GetLanguageById(ctx context.Context, req gen.GetLanguageByIdRequestObject) (gen.GetLanguageByIdResponseObject, error) {
	l, err := s.services.Languages.GetLanguage(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetLanguageById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetLanguageById500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetLanguageById200JSONResponse(toLanguageResponse(l)), nil
}

func (s *shared) CreateLanguage(ctx context.Context, req gen.CreateLanguageRequestObject) (gen.CreateLanguageResponseObject, error) {
	l, err := s.services.Languages.CreateLanguage(fromCreateLanguageRequest(req.Body))
	if err != nil {
		return gen.CreateLanguage400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateLanguage200JSONResponse(toLanguageResponse(l)), nil
}

func (s *shared) UpdateLanguage(ctx context.Context, req gen.UpdateLanguageRequestObject) (gen.UpdateLanguageResponseObject, error) {
	l, err := s.services.Languages.UpdateLanguage(uint(req.Id), fromUpdateLanguageRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateLanguage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateLanguage400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateLanguage200JSONResponse(toLanguageResponse(l)), nil
}

func (s *shared) DeleteLanguage(ctx context.Context, req gen.DeleteLanguageRequestObject) (gen.DeleteLanguageResponseObject, error) {
	if err := s.services.Languages.DeleteLanguage(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteLanguage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteLanguage404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteLanguage200Response{}, nil
}
