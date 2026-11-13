package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

func (s *shared) GetCards(ctx context.Context, req gen.GetCardsRequestObject) (gen.GetCardsResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	cards, total, err := s.services.Cards.GetCardsPublic(limit, offset)
	if err != nil {
		return gen.GetCards500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.CardResponse, 0, len(cards))
	for _, c := range cards {
		out = append(out, toCardResponse(c))
	}
	return gen.GetCards200JSONResponse{Data: &out, Meta: newListMeta(limit, offset, total)}, nil
}

func (s *shared) SearchCards(ctx context.Context, req gen.SearchCardsRequestObject) (gen.SearchCardsResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	cards, total, err := s.services.Cards.SearchCardsPublic(
		int(derefUint32(req.Params.GameId)),
		int(derefUint32(req.Params.LanguageId)),
		int(derefUint32(req.Params.PackId)),
		derefStr(req.Params.Query),
		limit, offset,
	)
	if err != nil {
		return gen.SearchCards500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.CardResponse, 0, len(cards))
	for _, c := range cards {
		out = append(out, toCardResponse(c))
	}
	return gen.SearchCards200JSONResponse{Data: &out, Meta: newListMeta(limit, offset, total)}, nil
}

func (s *shared) GetCardById(ctx context.Context, req gen.GetCardByIdRequestObject) (gen.GetCardByIdResponseObject, error) {
	c, err := s.services.Cards.GetCardPublic(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetCardById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetCardById500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetCardById200JSONResponse{Data: ptrTo(toCardResponse(c))}, nil
}

func (s *shared) CreateCard(ctx context.Context, req gen.CreateCardRequestObject) (gen.CreateCardResponseObject, error) {
	// The contract defines no 500 for CreateCard; map internal errors to 400.
	c, err := s.services.Cards.CreateCard(fromCreateCardRequest(req.Body))
	if err != nil {
		return gen.CreateCard400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.CreateCard200JSONResponse{Data: ptrTo(toCardResponse(c))}, nil
}

func (s *shared) UpdateCard(ctx context.Context, req gen.UpdateCardRequestObject) (gen.UpdateCardResponseObject, error) {
	c, err := s.services.Cards.UpdateCard(uint(req.Id), fromUpdateCardRequest(req.Body))
	if err != nil {
		if isNotFound(err) {
			return gen.UpdateCard404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		// The contract defines no 500 for UpdateCard; map internal errors to 400.
		return gen.UpdateCard400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateCard200JSONResponse{Data: ptrTo(toCardResponse(c))}, nil
}

func (s *shared) GetProtectedCards(ctx context.Context, req gen.GetProtectedCardsRequestObject) (gen.GetProtectedCardsResponseObject, error) {
	limit, offset := listParams(req.Params.Limit, req.Params.Offset)
	cards, total, err := s.services.Cards.GetCardsProtected(limit, offset)
	if err != nil {
		return gen.GetProtectedCards500JSONResponse{ServerErrorJSONResponse: gen.ServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	out := make([]gen.CardProtectedResponse, 0, len(cards))
	for _, c := range cards {
		out = append(out, toCardProtectedResponse(c))
	}
	return gen.GetProtectedCards200JSONResponse{Data: &out, Meta: newListMeta(limit, offset, total)}, nil
}

func (s *shared) GetProtectedCardById(ctx context.Context, req gen.GetProtectedCardByIdRequestObject) (gen.GetProtectedCardByIdResponseObject, error) {
	c, err := s.services.Cards.GetCardProtected(uint(req.Id))
	if err != nil {
		if isNotFound(err) {
			return gen.GetProtectedCardById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
		}
		return gen.GetProtectedCardById404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.GetProtectedCardById200JSONResponse{Data: ptrTo(toCardProtectedResponse(c))}, nil
}

func (s *shared) UpdateProtectedCard(ctx context.Context, req gen.UpdateProtectedCardRequestObject) (gen.UpdateProtectedCardResponseObject, error) {
	c, err := s.services.Cards.UpdateCard(uint(req.Id), fromUpdateCardRequest(req.Body))
	if err != nil {
		return gen.UpdateProtectedCard404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return gen.UpdateProtectedCard200JSONResponse{Data: ptrTo(toCardProtectedResponseFromCard(c))}, nil
}
