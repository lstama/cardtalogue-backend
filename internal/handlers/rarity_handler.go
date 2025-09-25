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
