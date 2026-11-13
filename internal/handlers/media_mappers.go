package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toMediaResponse(m *models.MediaResponse) gen.MediaResponse {
	if m == nil {
		return gen.MediaResponse{}
	}
	return gen.MediaResponse{
		Id:          uint32Ptr(uint32(m.ID)),
		Url:         strPtr(m.URL),
		Description: strPtr(m.Description),
		Status:      mediaStatusPtr(m.Status),
		Provider:    strPtr(m.Provider),
		CreatedAt:   timePtr(m.CreatedAt),
		UpdatedAt:   timePtr(m.UpdatedAt),
	}
}

func toMediaResponseFromMedia(m *models.Media) gen.MediaResponse {
	if m == nil {
		return gen.MediaResponse{}
	}
	return gen.MediaResponse{
		Id:          uint32Ptr(uint32(m.ID)),
		Url:         strPtr(m.URL),
		Description: strPtr(m.Description),
		Status:      mediaStatusPtr(m.Status),
		Provider:    strPtr(m.Provider),
		CreatedAt:   timePtr(m.CreatedAt),
		UpdatedAt:   timePtr(m.UpdatedAt),
	}
}

func toMediaScanResponse(r *models.CardScanResponse) gen.MediaScanResponse {
	if r == nil {
		return gen.MediaScanResponse{}
	}
	matches := make([]gen.MediaScanMatch, 0, len(r.Matches))
	for _, m := range r.Matches {
		matches = append(matches, gen.MediaScanMatch{
			Name:       strPtr(m.Name),
			Confidence: float32Ptr(float32(m.Confidence)),
			Detail: &struct {
				Effect         *string `json:"effect,omitempty"`
				MonsterEffect  *string `json:"monster_effect,omitempty"`
				PendulumEffect *string `json:"pendulum_effect,omitempty"`
			}{
				Effect:         m.Detail.Effect,
				MonsterEffect:  m.Detail.MonsterEffect,
				PendulumEffect: m.Detail.PendulumEffect,
			},
			KonamiId: strPtr(m.KonamiID),
		})
	}
	return gen.MediaScanResponse{
		Matches:        &matches,
		ProcessingTime: float32Ptr(float32(r.ProcessingTime)),
		Success:        boolPtr(r.Success),
	}
}

func toMediaUploadResponse(m *models.MediaResponse) gen.MediaUploadResponse {
	if m == nil {
		return gen.MediaUploadResponse{}
	}
	return gen.MediaUploadResponse{
		Id:  uint32Ptr(uint32(m.ID)),
		Url: strPtr(m.URL),
	}
}

func fromUpdateMediaRequest(r *gen.UpdateMediaRequest) *models.UpdateMediaRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateMediaRequest{
		URL:         r.Url,
		Description: r.Description,
		Status:      derefUpdateMediaStatusPtr(r.Status),
		Provider:    r.Provider,
	}
}
