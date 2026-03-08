package repositories

import (
	"backend/internal/models"
	"gorm.io/gorm"
	"time"
)

type GameFilter struct {
	Status   string    `form:"status"`
	GameMode string    `form:"game_mode"`
	MapID    string    `form:"map_id"`
	FromDate time.Time `form:"from_date" time_format:"2006-01-02T15:04:05Z07:00"`
	ToDate   time.Time `form:"to_date" time_format:"2006-01-02T15:04:05Z07:00"`
	Page     int       `form:"page,default=1"`
	Limit    int       `form:"limit,default=20"`
}

type GameRepository interface {
	Create(game *models.Game) error
	GetAll(filter GameFilter) ([]models.Game, int64, error)
	GetByID(id uint) (*models.Game, error)
	Update(game *models.Game) error
	Delete(id uint) error
	GetActive(page, limit int) ([]models.Game, error)
	GetByDateRange(from, to time.Time) ([]models.Game, error)
	UpdateStatus(id uint, status models.GameStatus) error
	UpdatePlayerCount(id uint, currentPlayers int) error
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) GetAll(filter GameFilter) ([]models.Game, int64, error) {
	var games []models.Game
	var total int64

	offset := (filter.Page - 1) * filter.Limit

	query := r.db.Model(&models.Game{})

	// Применяем фильтры
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.GameMode != "" {
		query = query.Where("game_mode = ?", filter.GameMode)
	}

	if filter.MapID != "" {
		query = query.Where("map_id = ?", filter.MapID)
	}

	if !filter.FromDate.IsZero() {
		query = query.Where("created_at >= ?", filter.FromDate)
	}

	if !filter.ToDate.IsZero() {
		query = query.Where("created_at <= ?", filter.ToDate)
	}

	// Считаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	if err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&games).Error; err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *gameRepository) GetByID(id uint) (*models.Game, error) {
	var game models.Game
	if err := r.db.First(&game, id).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) Update(game *models.Game) error {
	return r.db.Save(game).Error
}

func (r *gameRepository) Delete(id uint) error {
	return r.db.Delete(&models.Game{}, id).Error
}

func (r *gameRepository) GetActive(page, limit int) ([]models.Game, error) {
	var games []models.Game

	offset := (page - 1) * limit

	if err := r.db.Where("status = ?",
		models.StatusActive).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error; err != nil {
		return nil, err
	}

	return games, nil
}

func (r *gameRepository) GetByDateRange(from, to time.Time) ([]models.Game, error) {
	var games []models.Game

	if err := r.db.Where("created_at > ? ", from).
		Order("created_at ASC").
		Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (r *gameRepository) UpdateStatus(id uint, status models.GameStatus) error {
	return r.db.Model(&models.Game{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *gameRepository) UpdatePlayerCount(id uint, currentPlayers int) error {
	return r.db.Model(&models.Game{}).
		Where("id = ?", id).
		Update("current_players", currentPlayers).Error
}
