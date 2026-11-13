package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetGames(ctx context.Context, req gen.GetGamesRequestObject) (gen.GetGamesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	games, _, err := s.services.Games.GetGames(limit, offset)
	if err != nil {
		return gen.GetGames500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.GameResponse, 0, len(games))
	for _, g := range games {
		out = append(out, toGameResponse(g))
	}
	return gen.GetGames200JSONResponse(out), nil
}

func (s *shared) SearchGames(ctx context.Context, req gen.SearchGamesRequestObject) (gen.SearchGamesResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	games, _, err := s.services.Games.SearchGames(derefStr(req.Params.Query), limit, offset)
	if err != nil {
		return gen.SearchGames500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.GameResponse, 0, len(games))
	for _, g := range games {
		out = append(out, toGameResponse(g))
	}
	return gen.SearchGames200JSONResponse(out), nil
}

func (s *shared) GetGameById(ctx context.Context, req gen.GetGameByIdRequestObject) (gen.GetGameByIdResponseObject, error) {
	g, err := s.services.Games.GetGame(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetGameById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetGameById500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetGameById200JSONResponse(toGameResponse(g)), nil
}

func (s *shared) CreateGame(ctx context.Context, req gen.CreateGameRequestObject) (gen.CreateGameResponseObject, error) {
	g, err := s.services.Games.CreateGame(fromCreateGameRequest(req.Body))
	if err != nil {
		return gen.CreateGame400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateGame200JSONResponse(toGameResponse(g)), nil
}

func (s *shared) UpdateGame(ctx context.Context, req gen.UpdateGameRequestObject) (gen.UpdateGameResponseObject, error) {
	g, err := s.services.Games.UpdateGame(uint(req.Id), fromUpdateGameRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateGame404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.UpdateGame400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateGame200JSONResponse(toGameResponse(g)), nil
}

func (s *shared) DeleteGame(ctx context.Context, req gen.DeleteGameRequestObject) (gen.DeleteGameResponseObject, error) {
	if err := s.services.Games.DeleteGame(uint(req.Id)); err != nil {
		if isNotFound(err) {
			return gen.DeleteGame404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.DeleteGame404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.DeleteGame200JSONResponse{Message: strPtr("game deleted")}, nil
}
