package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/models"
)

func toUserResponse(u *models.UserResponse) gen.UserResponse {
	if u == nil {
		return gen.UserResponse{}
	}
	return gen.UserResponse{
		Id:        uint32Ptr(uint32(u.ID)),
		Name:      strPtr(u.Name),
		Email:     emailPtr(u.Email),
		Role:      userRolePtr(u.Role),
		Age:       intPtr(u.Age),
		CreatedAt: timePtr(u.CreatedAt),
		UpdatedAt: timePtr(u.UpdatedAt),
	}
}

func fromCreateUserRequest(r *gen.CreateUserRequest) *models.CreateUserRequest {
	if r == nil {
		return nil
	}
	return &models.CreateUserRequest{
		Name:     r.Name,
		Email:    string(r.Email),
		Password: r.Password,
		Role:     string(r.Role),
		Age:      r.Age,
	}
}

func fromUpdateUserRequest(r *gen.UpdateUserRequest) *models.UpdateUserRequest {
	if r == nil {
		return nil
	}
	return &models.UpdateUserRequest{
		Name:  r.Name,
		Email: derefEmailPtr(r.Email),
		Role:  derefUserRolePtr(r.Role),
		Age:   r.Age,
	}
}
