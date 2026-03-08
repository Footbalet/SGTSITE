package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/repositories"
	"backend/internal/services"
)

type GameHandler struct {
	gameService services.GameService
}

func NewGameHandler(db *gorm.DB) *GameHandler {
	gameRepo := repositories.NewGameRepository(db)
	gameService := services.NewGameService(gameRepo)
	return &GameHandler{gameService: gameService}
}

// CreateGame - POST /api/v1/games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req dto.CreateGameRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.CreateGame(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Game created successfully",
		"ID":      game.ID,
	})
}

// GetAllGames - GET /api/v1/games
func (h *GameHandler) GetAllGames(c *gin.Context) {
	var filter repositories.GameFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	games, total, err := h.gameService.GetAllGames(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": (int(total) + filter.Limit - 1) / filter.Limit,
		"results":     games,
	})
}

// GetGameByID - GET /api/v1/games/:id
func (h *GameHandler) GetGameByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	game, err := h.gameService.GetGameByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, game)
}

// UpdateGame - PUT /api/v1/games/:id
func (h *GameHandler) UpdateGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req dto.UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.UpdateGame(uint(id), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game updated successfully",
		"game":    game,
	})
}

// DeleteGame - DELETE /api/v1/games/:id
func (h *GameHandler) DeleteGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.gameService.DeleteGame(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game deleted successfully",
	})
}

// GetActiveGames - GET /api/v1/games/active
func (h *GameHandler) GetActiveGames(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	games, err := h.gameService.GetActiveGames(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(games),
		"results": games,
	})
}

// FinishGame - POST /api/v1/games/:id/finish
func (h *GameHandler) FinishGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	game, err := h.gameService.FinishGame(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game finished successfully",
		"game":    game,
	})
}

// CancelGame - POST /api/v1/games/:id/cancel
func (h *GameHandler) CancelGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	game, err := h.gameService.CancelGame(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game cancelled successfully",
		"game":    game,
	})
}

// UpdatePlayerCount - PATCH /api/v1/games/:id/players
func (h *GameHandler) UpdatePlayerCount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		CurrentPlayers int `json:"current_players" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.UpdatePlayerCount(uint(id), req.CurrentPlayers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player count updated successfully",
		"game":    game,
	})
}

// GetGamesStats - GET /api/v1/games/stats
func (h *GameHandler) GetGamesStats(c *gin.Context) {
	// Пример статистики
	from := time.Now().AddDate(-2, -1, 0)
	to := time.Now().AddDate(0, 1, 0)

	games, err := h.gameService.GetGamesByDateRange(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Рассчитываем статистику
	var totalGames, activeGames, finishedGames, cancelledGames int
	var totalPlayers int

	gameModeCount := make(map[string]int)
	mapIDCount := make(map[string]int)

	for _, game := range games {
		totalGames++
		switch game.Status {
		case "active":
			activeGames++
		case "finished":
			finishedGames++
		case "cancelled":
			cancelledGames++
		}
		totalPlayers += game.MaxPlayers

		gameModeCount[string(game.GameMode)]++
		mapIDCount[game.MapID]++
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"totalStats": gin.H{
			"total_games":     totalGames,
			"active_games":    activeGames,
			"finished_games":  finishedGames,
			"cancelled_games": cancelledGames,
			"total_players":   totalPlayers,
			"avg_players_per_game": func() float64 {
				if totalGames > 0 {
					return float64(totalPlayers) / float64(totalGames)
				}
				return 0
			}(),
		},
		"modesStats": gameModeCount,
		"mapsStats":  mapIDCount,
	})
}
