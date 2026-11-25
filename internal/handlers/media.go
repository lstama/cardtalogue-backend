package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) ScanMedia(ctx context.Context, req gen.ScanMediaRequestObject) (gen.ScanMediaResponseObject, error) {
	// Proof-of-work captcha (preserved from the old route), now driven by the
	// typed X-Cap-Token header parameter instead of a gin context assertion.
	if msg := s.verifyCaptcha(ctx, req.Params.XCapToken); msg != "" {
		return gen.ScanMedia400JSONResponse{Error: msg}, nil
	}

	fh, err := extractFileHeader(req.Body)
	if err != nil {
		return gen.ScanMedia400JSONResponse{Error: err.Error()}, nil
	}
	resp, err := s.services.Media.ScanImage(fh)
	if err != nil {
		return gen.ScanMedia400JSONResponse{Error: err.Error()}, nil
	}
	return gen.ScanMedia200JSONResponse{Data: ptrTo(toMediaScanResponse(resp))}, nil
}

func (s *shared) UploadMedia(ctx context.Context, req gen.UploadMediaRequestObject) (gen.UploadMediaResponseObject, error) {
	fh, err := extractFileHeader(req.Body)
	if err != nil {
		return gen.UploadMedia400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	media, _, err := s.services.Media.UploadImage(fh, s.cfg.APIBaseURL, s.cfg.MediaUploadPath, true)
	if err != nil {
		return gen.UploadMedia400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UploadMedia200JSONResponse{Data: ptrTo(toMediaUploadResponse(media))}, nil
}

func (s *shared) GetMediaById(ctx context.Context, req gen.GetMediaByIdRequestObject) (gen.GetMediaByIdResponseObject, error) {
	m, err := s.services.Media.GetMedia(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetMediaById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetMediaById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetMediaById200JSONResponse{Data: ptrTo(toMediaResponseFromMedia(m))}, nil
}

func (s *shared) UpdateMedia(ctx context.Context, req gen.UpdateMediaRequestObject) (gen.UpdateMediaResponseObject, error) {
	m, err := s.services.Media.UpdateMedia(uint(req.Id), fromUpdateMediaRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateMedia404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateMedia400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateMedia200JSONResponse{Data: ptrTo(toMediaResponse(m))}, nil
}

func (s *shared) DeleteMedia(ctx context.Context, req gen.DeleteMediaRequestObject) (gen.DeleteMediaResponseObject, error) {
	if err := s.services.Media.DeleteMedia(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteMedia404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteMedia404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteMedia200Response{}, nil
}
