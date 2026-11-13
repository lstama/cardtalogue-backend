package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) ScanMedia(ctx context.Context, req gen.ScanMediaRequestObject) (gen.ScanMediaResponseObject, error) {
	// Proof-of-Work captcha (preserved from the old route). ctx is the gin context
	// at runtime; verify the token header before processing.
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if s.verifier.Enabled() {
			token := ginCtx.GetHeader("X-Cap-Token")
			if token == "" {
				return gen.ScanMedia400JSONResponse{Error: "captcha required"}, nil
			}
			ok, err := s.verifier.Verify(ginCtx.Request.Context(), token)
			if err != nil {
				return gen.ScanMedia400JSONResponse{Error: "captcha unavailable"}, nil
			}
			if !ok {
				return gen.ScanMedia400JSONResponse{Error: "captcha failed"}, nil
			}
		}
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
