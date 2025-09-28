package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	gameService service.GameService
}

func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// CreateGame godoc
// @Summary Create a new game
// @Description Create a new game with the provided details
// @Tags Games
// @Accept json
// @Produce json
// @Param request body models.CreateGameRequest true "Game details"
// @Security BearerAuth
// @Success 201 {object} object{data=models.GameResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	game, err := h.gameService.CreateGame(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": game})
}

// GetGame godoc
// @Summary Get game by ID
// @Description Retrieve a game by its ID
// @Tags Games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} object{data=models.GameResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{id} [get]
func (h *GameHandler) GetGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}
	game, err := h.gameService.GetGame(uint(id))
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": game})
}

// GetGames godoc
// @Summary Get list of games
// @Description Retrieve a list of games with pagination
// @Tags Games
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.GameResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games [get]
func (h *GameHandler) GetGames(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	games, total, err := h.gameService.GetGames(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": games,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateGame godoc
// @Summary Update game by ID
// @Description Update a game's details by its ID
// @Tags Games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Param request body models.UpdateGameRequest true "Updated game details"
// @Security BearerAuth
// @Success 200 {object} object{data=models.GameResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games/{id} [put]
func (h *GameHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}
	var req models.UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	game, err := h.gameService.UpdateGame(uint(id), &req)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": game})
}

// DeleteGame godoc
// @Summary Delete game by ID
// @Description Delete a game by its ID
// @Tags Games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Security BearerAuth
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games/{id} [delete]
func (h *GameHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}
	if err := h.gameService.DeleteGame(uint(id)); err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "game deleted successfully"})
}

// SearchGames godoc
// @Summary Search games
// @Description Search for games by name with pagination
// @Tags Games
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.GameResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games/search [get]
func (h *GameHandler) SearchGames(c *gin.Context) {
	query := c.Query("query")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	games, total, err := h.gameService.SearchGames(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": games,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}
