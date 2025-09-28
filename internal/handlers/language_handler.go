package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type LanguageHandler struct {
	languageService service.LanguageService
}

func NewLanguageHandler(languageService service.LanguageService) *LanguageHandler {
	return &LanguageHandler{languageService: languageService}
}

// CreateLanguage godoc
// @Summary Create a new language
// @Description Create a new language with the provided details
// @Tags Languages
// @Accept json
// @Produce json
// @Param request body models.CreateLanguageRequest true "Language details"
// @Security BearerAuth
// @Success 201 {object} object{data=models.LanguageResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /languages [post]
func (h *LanguageHandler) CreateLanguage(c *gin.Context) {
	var req models.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	language, err := h.languageService.CreateLanguage(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": language})
}

// GetLanguage godoc
// @Summary Get language by ID
// @Description Retrieve a language by its ID
// @Tags Languages
// @Accept json
// @Produce json
// @Param id path int true "Language ID"
// @Success 200 {object} object{data=models.LanguageResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /languages/{id} [get]
func (h *LanguageHandler) GetLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid language ID"})
		return
	}
	language, err := h.languageService.GetLanguage(uint(id))
	if err != nil {
		if err.Error() == "language not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": language})
}

// GetLanguages godoc
// @Summary Get all languages
// @Description Retrieve a list of all languages with pagination
// @Tags Languages
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.LanguageResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 500 {object} map[string]string
// @Router /languages [get]
func (h *LanguageHandler) GetLanguages(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	languages, total, err := h.languageService.GetLanguages(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": languages,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateLanguage godoc
// @Summary Update a language
// @Description Update the details of an existing language by its ID
// @Tags Languages
// @Accept json
// @Produce json
// @Param id path int true "Language ID"
// @Param request body models.UpdateLanguageRequest true "Updated language details"
// @Security BearerAuth
// @Success 200 {object} object{data=models.LanguageResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /languages/{id} [put]
func (h *LanguageHandler) UpdateLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid language ID"})
		return
	}
	var req models.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	language, err := h.languageService.UpdateLanguage(uint(id), &req)
	if err != nil {
		if err.Error() == "language not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": language})
}

// DeleteLanguage godoc
// @Summary Delete a language
// @Description Delete an existing language by its ID
// @Tags Languages
// @Accept json
// @Produce json
// @Param id path int true "Language ID"
// @Success 200 {object} object{message=string}
// @Security BearerAuth
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /languages/{id} [delete]
func (h *LanguageHandler) DeleteLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid language ID"})
		return
	}
	if err := h.languageService.DeleteLanguage(uint(id)); err != nil {
		if err.Error() == "language not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "language deleted successfully"})
}

// SearchLanguages godoc
// @Summary Search languages
// @Description Search for languages by name or code with pagination
// @Tags Languages
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.LanguageResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /languages/search [get]
func (h *LanguageHandler) SearchLanguages(c *gin.Context) {
	query := c.Query("query")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	languages, total, err := h.languageService.SearchLanguages(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": languages,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}
