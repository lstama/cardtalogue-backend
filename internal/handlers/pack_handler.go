package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PackHandler struct {
	packService service.PackService
}

func NewPackHandler(packService service.PackService) *PackHandler {
	return &PackHandler{packService: packService}
}

// CreatePack godoc
// @Summary Create a new pack
// @Description Create a new pack with the provided details
// @Tags Packs
// @Accept json
// @Produce json
// @Param request body models.CreatePackRequest true "Pack details"
// @Success 201 {object} object{data=models.PackResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /packs [post]
func (h *PackHandler) CreatePack(c *gin.Context) {
	var req models.CreatePackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pack, err := h.packService.CreatePack(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": pack})
}

// GetPack godoc
// @Summary Get pack by ID
// @Description Retrieve a pack by its ID
// @Tags Packs
// @Accept json
// @Produce json
// @Param id path int true "Pack ID"
// @Success 200 {object} object{data=models.PackResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /packs/{id} [get]
func (h *PackHandler) GetPack(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pack ID"})
		return
	}

	pack, err := h.packService.GetPack(uint(id))
	if err != nil {
		if err.Error() == "pack not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pack})
}

// GetPacks godoc
// @Summary Get list of packs
// @Description Retrieve a list of packs with pagination
// @Tags Packs
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.PackResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /packs [get]
func (h *PackHandler) GetPacks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	packs, total, err := h.packService.GetPacks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": packs,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdatePack godoc
// @Summary Update a pack
// @Description Update the details of an existing pack
// @Tags Packs
// @Accept json
// @Produce json
// @Param id path int true "Pack ID"
// @Param request body models.UpdatePackRequest true "Updated pack details"
// @Success 200 {object} object{data=models.PackResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /packs/{id} [put]
func (h *PackHandler) UpdatePack(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pack ID"})
		return
	}

	var req models.UpdatePackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pack, err := h.packService.UpdatePack(uint(id), &req)
	if err != nil {
		if err.Error() == "pack not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pack})
}

// DeletePack godoc
// @Summary Delete a pack
// @Description Delete a pack by its ID
// @Tags Packs
// @Accept json
// @Produce json
// @Param id path int true "Pack ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /packs/{id} [delete]
func (h *PackHandler) DeletePack(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pack ID"})
		return
	}

	if err := h.packService.DeletePack(uint(id)); err != nil {
		if err.Error() == "pack not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pack deleted successfully"})
}

// SearchPacks godoc
// @Summary Search packs
// @Description Search for packs based on game ID, language ID, and query string with pagination, use id = 0 to ignore a filter
// @Tags Packs
// @Accept json
// @Produce json
// @Param game_id query int false "Game ID"
// @Param language_id query int false "Language ID"
// @Param query query string false "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.PackResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /packs/search [get]
func (h *PackHandler) SearchPacks(c *gin.Context) {
	gameID, _ := strconv.Atoi(c.Query("game_id"))
	languageID, _ := strconv.Atoi(c.Query("language_id"))
	queryString := c.Query("query")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	packs, total, err := h.packService.SearchPacks(gameID, languageID, queryString, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": packs,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}
