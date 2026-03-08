package handlers

import (
	"fmt"
	_ "io"
	"net/http"
	"os"
	_ "path/filepath"
	"strconv"
	_ "strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/repositories"
	"backend/internal/services"
)

type ReleaseHandler struct {
	releaseService services.ReleaseService
}

func NewReleaseHandler(db *gorm.DB, uploadDir, baseURL string) *ReleaseHandler {
	releaseRepo := repositories.NewReleaseRepository(db)
	releaseService := services.NewReleaseService(releaseRepo, uploadDir, baseURL)
	return &ReleaseHandler{releaseService: releaseService}
}

// CreateRelease - POST /api/v1/releases
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	var req dto.CreateReleaseRequest

	authHeader := c.GetHeader("Authorization")
	if authHeader != "jwflasfjasopkfaroqjakl;jddopaqjdakjqio" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
		return
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	release, err := h.releaseService.CreateRelease(
		req.Version,
		req.Platform,
		req.Changes,
		"132dsd",
		true,
		req.File,
	)
	print("999999999999")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Release created successfully",
		"release": release,
	})
}

// GetAllReleases - GET /api/v1/releases
func (h *ReleaseHandler) GetAllReleases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	platform := c.Query("platform")
	sortIndex := c.Query("sortIndex")
	if platform == "0" {
		platform = ""
	}
	if platform == "1" {
		platform = "Windows"
	}
	if platform == "2" {
		platform = "Linux"
	}
	releases, total, err := h.releaseService.GetAllReleases(page, limit, platform, sortIndex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       total,
		"page":        page,
		"limit":       limit,
		"platform":    platform,
		"total_pages": (int(total) + limit - 1) / limit,
		"results":     releases,
	})
}

// GetReleaseByID - GET /api/v1/releases/:id
func (h *ReleaseHandler) GetReleaseByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	release, err := h.releaseService.GetReleaseByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, release)
}

// GetLatestRelease - GET /api/v1/releases/latest
func (h *ReleaseHandler) GetLatestRelease(c *gin.Context) {
	platform := c.Query("platform")

	release, err := h.releaseService.GetLatestRelease(platform)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active releases found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, release)
}

// UpdateRelease - PUT /api/v1/releases/:id
func (h *ReleaseHandler) UpdateRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "jwflasfjasopkfaroqjakl;jddopaqjdakjqio" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
		return
	}
	var req dto.UpdateReleaseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	release, err := h.releaseService.UpdateRelease(
		uint(id),
		req.Version,
		req.Platform,
		req.Changes,
		"sasdsdsd",
		true,
		req.File,
	)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Release updated successfully",
		"release": release,
	})
}

// DeleteRelease - DELETE /api/v1/releases/:id
func (h *ReleaseHandler) DeleteRelease(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "jwflasfjasopkfaroqjakl;jddopaqjdakjqio" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.releaseService.DeleteRelease(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Release deleted successfully",
	})
}

// DownloadRelease - GET /api/v1/releases/:id/download
func (h *ReleaseHandler) DownloadRelease(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ip := c.ClientIP()

	release, filePath, err := h.releaseService.DownloadRelease(uint(id), ip)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found"})
		return
	}
	defer file.Close()

	// Получаем статистику файла
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot get file info"})
		return
	}

	// Устанавливаем заголовки для скачивания
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", release.Filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("X-File-Size", fmt.Sprintf("%d", release.FileSize))
	c.Header("X-Checksum", release.Checksum)
	c.Header("X-Version", release.Version)

	// Отправляем файл
	c.File(filePath)
}
