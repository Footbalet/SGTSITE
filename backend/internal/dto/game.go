package dto

import (
	"backend/internal/models"
	"time"
)

type CreateGameRequest struct {
	MapID          string            `json:"map_id" binding:"required,min=1,max=50"`
	GameMode       models.GameMode   `json:"game_mode" binding:"required"`
	Status         models.GameStatus `json:"status" binding:"omitempty,oneof=scheduled active finished cancelled"`
	CurrentPlayers int               `json:"players" binding:"omitempty,min=1,max=64"`
}

type UpdateGameRequest struct {
	Finish         *time.Time        `json:"finish"`
	MapID          string            `json:"map_id" binding:"omitempty,min=1,max=50"`
	GameMode       models.GameMode   `json:"game_mode" binding:"omitempty"`
	Status         models.GameStatus `json:"status" binding:"omitempty,oneof=scheduled active finished cancelled"`
	ServerIP       string            `json:"server_ip" binding:"omitempty,ipv4"`
	MaxPlayers     *int              `json:"max_players" binding:"omitempty,min=2,max=64"`
	CurrentPlayers *int              `json:"current_players" binding:"omitempty,min=0,max=64"`
}

type GameResponse struct {
	ID             uint              `json:"id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Finish         time.Time         `json:"finish"`
	MapID          string            `json:"map_id"`
	GameMode       models.GameMode   `json:"game_mode"`
	Status         models.GameStatus `json:"status"`
	ServerIP       string            `json:"server_ip,omitempty"`
	MaxPlayers     int               `json:"max_players"`
	CurrentPlayers int               `json:"current_players"`
	Duration       int               `json:"duration"`
	IsActive       bool              `json:"is_active"`
	IsUpcoming     bool              `json:"is_upcoming"`
}
