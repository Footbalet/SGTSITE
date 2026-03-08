package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/models"
	"backend/internal/repositories"
	"backend/internal/services"
)

type NewsHandler struct {
	newsService services.NewsService
}

func NewNewsHandler(db *gorm.DB) *NewsHandler {
	newsRepo := repositories.NewNewsRepository(db)
	newsService := services.NewNewsService(newsRepo)
	return &NewsHandler{newsService: newsService}
}

// CreateNews - POST /api/v1/news
func (h *NewsHandler) CreateNews(c *gin.Context) {
	var input models.News

	authHeader := c.GetHeader("Authorization")
	if authHeader != "jwflasfjasopkfaroqjakl;jddopaqjdakjqio" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.newsService.CreateNews(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "News created successfully",
		"news":    input,
	})
}

// GetAllNews - GET /api/v1/news
func (h *NewsHandler) GetAllNews(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	news, total, err := h.newsService.GetAllNews(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
		"results":     news,
	})
}

// GetNewsByID - GET /api/v1/news/:id
func (h *NewsHandler) GetNewsByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	news, err := h.newsService.GetNewsByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, news)
}

// UpdateNews - PUT /api/v1/news/:id
func (h *NewsHandler) UpdateNews(c *gin.Context) {

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

	var input models.News
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.newsService.UpdateNews(uint(id), &input); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "News updated successfully",
	})
}

// DeleteNews - DELETE /api/v1/news/:id
func (h *NewsHandler) DeleteNews(c *gin.Context) {
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

	if err := h.newsService.DeleteNews(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "News deleted successfully",
	})
}

// GetNewsByTheme - GET /api/v1/news/theme/:theme
func (h *NewsHandler) GetNewsByTheme(c *gin.Context) {
	theme := c.Param("theme")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	news, total, err := h.newsService.GetNewsByTheme(theme, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       total,
		"page":        page,
		"limit":       limit,
		"theme":       theme,
		"total_pages": (int(total) + limit - 1) / limit,
		"results":     news,
	})
}

// SearchNews - GET /api/v1/news/search
func (h *NewsHandler) SearchNews(c *gin.Context) {
	theme_code, _ := strconv.Atoi(c.DefaultQuery("theme", "0"))
	sortIndex := c.DefaultQuery("sortIndex", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	news, total, err := h.newsService.SearchNews(theme_code, sortIndex, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       total,
		"page":        page,
		"limit":       limit,
		"query":       theme_code,
		"total_pages": (int(total) + limit - 1) / limit,
		"results":     news,
	})
}

// GetNews - обработчик без зависимости от сервиса
func GetNews(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		newsHandler := NewNewsHandler(db)
		newsHandler.GetAllNews(c)
	}
}

// GetNewsByID - обработчик без зависимости от сервиса
func GetNewsByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		newsHandler := NewNewsHandler(db)
		newsHandler.GetNewsByID(c)
	}
}
