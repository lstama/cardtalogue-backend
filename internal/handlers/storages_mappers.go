package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toStorageResponse(st *models.StorageResponse) gen.StorageResponse {
	if st == nil {
		return gen.StorageResponse{}
	}
	return gen.StorageResponse{
		Id:          uint32Ptr(uint32(st.ID)),
		Code:        strPtr(st.Code),
		Name:        strPtr(st.Name),
		Description: strPtr(st.Description),
		CreatedAt:   timePtr(st.CreatedAt),
		UpdatedAt:   timePtr(st.UpdatedAt),
	}
}

func fromCreateStorageRequest(r *gen.CreateStorageRequest) *models.CreateStorageRequest {
	if r == nil {
		return nil
	}
	return &models.CreateStorageRequest{
		Code:        r.Code,
		Name:        r.Name,
		Description: derefStr(r.Description),
	}
}

func fromUpdateStorageRequest(r *gen.UpdateStorageRequest) *models.UpdateStorageRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateStorageRequest{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
	}
}
