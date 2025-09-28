package handlers

import (
	"net/http"
	"strconv"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	storageService service.StorageService
}

func NewStorageHandler(storageService service.StorageService) *StorageHandler {
	return &StorageHandler{storageService: storageService}
}

// CreateStorage godoc
// @Summary Create a new storage
// @Description Create a new storage with the provided details
// @Tags Storages
// @Accept json
// @Produce json
// @Param request body models.CreateStorageRequest true "Storage details"
// @Security BearerAuth
// @Success 201 {object} object{data=models.StorageResponse}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /storages [post]
func (h *StorageHandler) CreateStorage(c *gin.Context) {
	var req models.CreateStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	storage, err := h.storageService.CreateStorage(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": storage})
}

// GetStorage godoc
// @Summary Get storage by ID
// @Description Retrieve a storage by its ID
// @Tags Storages
// @Accept json
// @Produce json
// @Param id path int true "Storage ID"
// @Security BearerAuth
// @Success 200 {object} object{data=models.StorageResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /storages/{id} [get]
func (h *StorageHandler) GetStorage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid storage ID"})
		return
	}
	storage, err := h.storageService.GetStorage(uint(id))
	if err != nil {
		if err.Error() == "storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": storage})
}

// GetStorages godoc
// @Summary Get list of storages
// @Description Retrieve a list of storages with pagination
// @Tags Storages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object{data=[]models.StorageResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 500 {object} map[string]string
// @Router /storages [get]
func (h *StorageHandler) GetStorages(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	storages, total, err := h.storageService.GetStorages(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": storages,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateStorage godoc
// @Summary Update storage by ID
// @Description Update an existing storage with the provided details
// @Tags Storages
// @Accept json
// @Produce json
// @Param id path int true "Storage ID"
// @Param request body models.UpdateStorageRequest true "Updated storage details"
// @Security BearerAuth
// @Success 200 {object} object{data=models.StorageResponse}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /storages/{id} [put]
func (h *StorageHandler) UpdateStorage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid storage ID"})
		return
	}
	var req models.UpdateStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	storage, err := h.storageService.UpdateStorage(uint(id), &req)
	if err != nil {
		if err.Error() == "storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": storage})
}

// DeleteStorage godoc
// @Summary Delete storage by ID
// @Description Delete an existing storage by its ID
// @Tags Storages
// @Accept json
// @Produce json
// @Param id path int true "Storage ID"
// @Security BearerAuth
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /storages/{id} [delete]
func (h *StorageHandler) DeleteStorage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid storage ID"})
		return
	}
	if err := h.storageService.DeleteStorage(uint(id)); err != nil {
		if err.Error() == "storage not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "storage deleted successfully"})
}

// SearchStorages godoc
// @Summary Search storages
// @Description Search for storages based on a query string with pagination
// @Tags Storages
// @Accept json
// @Produce json
// @Param query_string query string true "Query string"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Security BearerAuth
// @Success 200 {object} object{data=[]models.StorageResponse,meta=object{total=int,limit=int,offset=int}}
// @Failure 500 {object} map[string]string
// @Router /storages/search [get]
func (h *StorageHandler) SearchStorages(c *gin.Context) {
	queryString := c.Query("query_string")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	storages, total, err := h.storageService.SearchStorages(queryString, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": storages,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}
