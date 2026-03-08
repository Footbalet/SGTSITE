package models

import (
	"gorm.io/gorm"
	"time"
)

type GameStatus string

const (
	StatusActive    GameStatus = "active"
	StatusFinished  GameStatus = "finished"
	StatusCancelled GameStatus = "cancelled"
)

type GameMode string

const (
	ModeDeathmatch GameMode = "deathmatch"
	ModeLastHero   GameMode = "lasthero"
	ModeKingOfHill GameMode = "king_of_the_hill"
	ModeRacing     GameMode = "racing"

	ModeTeamDeath   GameMode = "team_deathmatch"
	ModeCaptureFlag GameMode = "capture_the_flag"
	ModeDominating  GameMode = "dominating"
	ModeHarvest     GameMode = "harvest"
)

type Game struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Finish   time.Time  `json:"finish" gorm:"not null"`                                      // Время окончания игры
	MapID    string     `json:"map_id" gorm:"not null;size:50"`                              // ID карты
	GameMode GameMode   `json:"game_mode" gorm:"not null;type:varchar(50)"`                  // Режим игры
	Status   GameStatus `json:"status" gorm:"not null;type:varchar(20);default:'scheduled'"` // Статус игры

	// Дополнительные поля
	ServerIP       string `json:"server_ip" gorm:"size:50"`         // IP сервера
	MaxPlayers     int    `json:"max_players" gorm:"default:24"`    // Максимум игроков
	CurrentPlayers int    `json:"current_players" gorm:"default:0"` // Текущее количество игроков
	Duration       int    `json:"duration" gorm:"default:0"`        // Длительность в минутах
}

// IsValid проверяет корректность данных игры
func (g *Game) IsValid() bool {
	return g.CreatedAt.Before(g.Finish)
}

// CalculateDuration вычисляет длительность игры
func (g *Game) CalculateDuration() {
	if !g.CreatedAt.IsZero() && !g.Finish.IsZero() {
		g.Duration = int(g.Finish.Sub(g.CreatedAt).Minutes())
	}
}
