package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetUsers(ctx context.Context, req gen.GetUsersRequestObject) (gen.GetUsersResponseObject, error) {
	limit, offset := listParams(nil, nil)
	users, _, err := s.services.Users.GetUsers(limit, offset)
	if err != nil {
		return gen.GetUsers500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}
	return gen.GetUsers200JSONResponse(out), nil
}

func (s *shared) GetUserById(ctx context.Context, req gen.GetUserByIdRequestObject) (gen.GetUserByIdResponseObject, error) {
	u, err := s.services.Users.GetUser(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetUserById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetUserById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetUserById200JSONResponse(toUserResponse(u)), nil
}

func (s *shared) CreateUser(ctx context.Context, req gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	u, err := s.services.Users.CreateUser(fromCreateUserRequest(req.Body))
	if err != nil {
		return gen.CreateUser400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateUser200JSONResponse(toUserResponse(u)), nil
}

func (s *shared) UpdateUser(ctx context.Context, req gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	u, err := s.services.Users.UpdateUser(uint(req.Id), fromUpdateUserRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateUser404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateUser400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateUser200JSONResponse(toUserResponse(u)), nil
}

func (s *shared) DeleteUser(ctx context.Context, req gen.DeleteUserRequestObject) (gen.DeleteUserResponseObject, error) {
	if err := s.services.Users.DeleteUser(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteUser404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteUser404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteUser200Response{}, nil
}
