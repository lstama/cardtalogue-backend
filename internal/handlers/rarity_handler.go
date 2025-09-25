package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RarityHandler struct {
	rarityService service.RarityService
}

func NewRarityHandler(rarityService service.RarityService) *RarityHandler {
	return &RarityHandler{rarityService: rarityService}
}

// CreateRarity godoc
// @Summary Create a new rarity
// @Description Create a new rarity with the provided details
// @Tags Rarities
// @Accept json
// @Produce json
// @Param request body models.CreateRarityRequest true "Rarity details"
// @Success 201 {object} object{data=models.RarityResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rarities [post]
func (h *RarityHandler) CreateRarity(c *gin.Context) {
	var req models.CreateRarityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rarity, err := h.rarityService.CreateRarity(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rarity})
}

// GetRarity godoc
// @Summary Get rarity by ID
// @Description Retrieve a rarity by its ID
// @Tags Rarities
// @Accept json
// @Produce json
// @Param id path int true "Rarity ID"
// @Success 200 {object} object{data=models.RarityResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /rarities/{id} [get]
func (h *RarityHandler) GetRarity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rarity ID"})
		return
	}
	rarity, err := h.rarityService.GetRarity(uint(id))
	if err != nil {
		if err.Error() == "rarity not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rarity})
}

// GetRarities godoc
// @Summary Get list of rarities
// @Description Retrieve a list of rarities with pagination
// @Tags Rarities
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.RarityResponse, meta=object{total=int, limit=int, offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rarities [get]
func (h *RarityHandler) GetRarities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	rarities, total, err := h.rarityService.GetRarities(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": rarities,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateRarity godoc
// @Summary Update a rarity
// @Description Update an existing rarity by its ID
// @Tags Rarities
// @Accept json
// @Produce json
// @Param id path int true "Rarity ID"
// @Param request body models.UpdateRarityRequest true "Updated rarity details"
// @Success 200 {object} object{data=models.RarityResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rarities/{id} [put]
func (h *RarityHandler) UpdateRarity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rarity ID"})
		return
	}
	var req models.UpdateRarityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rarity, err := h.rarityService.UpdateRarity(uint(id), &req)
	if err != nil {
		if err.Error() == "rarity not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rarity})
}

// DeleteRarity godoc
// @Summary Delete a rarity
// @Description Delete an existing rarity by its ID
// @Tags Rarities
// @Accept json
// @Produce json
// @Param id path int true "Rarity ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rarities/{id} [delete]
func (h *RarityHandler) DeleteRarity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rarity ID"})
		return
	}
	if err := h.rarityService.DeleteRarity(uint(id)); err != nil {
		if err.Error() == "rarity not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rarity deleted successfully"})
}

// SearchRarities godoc
// @Summary Search rarities
// @Description Search rarities by query string with pagination
// @Tags Rarities
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.RarityResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rarities/search [get]
func (h *RarityHandler) SearchRarities(c *gin.Context) {
	queryString := c.Query("query")
	if queryString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	rarities, err := h.rarityService.SearchRarities(queryString, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rarities})
}
