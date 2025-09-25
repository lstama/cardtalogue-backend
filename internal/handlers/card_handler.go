package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	cardService service.CardService
}

func NewCardHandler(cardService service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

// CreateCard godoc
// @Summary Create a new card
// @Description Create a new card with the provided details
// @Tags Cards
// @Accept json
// @Produce json
// @Param request body models.CreateCardRequest true "Card details"
// @Success 201 {object} object{data=models.CardResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards [post]
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req models.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardService.CreateCard(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": card})
}

// GetCard godoc
// @Summary Get card by ID
// @Description Retrieve a card by its ID
// @Tags Cards
// @Accept json
// @Produce json
// @Param id path int true "Card ID"
// @Success 200 {object} object{data=models.CardResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /cards/{id} [get]
func (h *CardHandler) GetCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	card, err := h.cardService.GetCard(uint(id))
	if err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": card})
}

// GetCardProtected godoc
// @Summary Get card by ID, with protected data that isn't available to public
// @Description Retrieve a card by its ID, including protected data
// @Tags Cards
// @Accept json
// @Produce json
// @Param id path int true "Card ID"
// @Success 200 {object} object{data=models.CardProtectedResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /cards/protected/{id} [get]
func (h *CardHandler) GetCardProtected(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	card, err := h.cardService.GetCardProtected(uint(id))
	if err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": card})
}

// GetCards godoc
// @Summary Get list of cards
// @Description Retrieve a list of cards with pagination
// @Tags Cards
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.CardResponse, meta=object{total=int, limit=int, offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards [get]
func (h *CardHandler) GetCards(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	cards, total, err := h.cardService.GetCards(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cards,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetCardsProtected godoc
// @Summary Get list of cards, with protected data that isn't available to public
// @Description Retrieve a list of cards with protected data and pagination
// @Tags Cards
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.CardProtectedResponse, meta=object{total=int, limit=int, offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards/protected [get]
func (h *CardHandler) GetCardsProtected(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	cards, total, err := h.cardService.GetCardsProtected(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cards,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateCard godoc
// @Summary Update card by ID
// @Description Update a card's details by its ID
// @Tags Cards
// @Accept json
// @Produce json
// @Param id path int true "Card ID"
// @Param request body models.UpdateCardRequest true "Updated card details"
// @Success 200 {object} object{data=models.CardResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards/{id} [put]
func (h *CardHandler) UpdateCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	var req models.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardService.UpdateCard(uint(id), &req)
	if err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": card})
}

// DeleteCard godoc
// @Summary Delete card by ID
// @Description Delete a card by its ID
// @Tags Cards
// @Accept json
// @Produce json
// @Param id path int true "Card ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards/{id} [delete]
func (h *CardHandler) DeleteCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	if err := h.cardService.DeleteCard(uint(id)); err != nil {
		if err.Error() == "card not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "card deleted successfully"})
}

// SearchCards godoc
// @Summary Search cards
// @Description Search for cards based on game, language, pack, and query string with pagination, use id = 0 to ignore a filter
// @Tags Cards
// @Accept json
// @Produce json
// @Param game_id query int false "Game ID"
// @Param language_id query int false "Language ID"
// @Param pack_id query int false "Pack ID"
// @Param query query string false "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.CardResponse, meta=object{limit=int, offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards/search [get]
func (h *CardHandler) SearchCards(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	game_id, _ := strconv.Atoi(c.Query("game_id"))
	language_id, _ := strconv.Atoi(c.Query("language_id"))
	pack_id, _ := strconv.Atoi(c.Query("pack_id"))
	query_string := c.Query("query")

	cards, err := h.cardService.SearchCards(game_id, language_id, pack_id, query_string, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cards,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}
