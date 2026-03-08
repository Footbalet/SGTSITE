package repositories

import (
	"backend/internal/models"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type NewsRepository interface {
	Create(news *models.News) error
	GetAll(page, limit int) ([]models.News, int64, error)
	GetByID(id uint) (*models.News, error)
	Update(news *models.News) error
	Delete(id uint) error
	GetByTheme(theme string, page, limit int) ([]models.News, int64, error)
	Search(query int, sortIndex string, page, limit int) ([]models.News, int64, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(news *models.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) GetAll(page, limit int) ([]models.News, int64, error) {
	var news []models.News
	var total int64

	offset := (page - 1) * limit

	// Получаем общее количество
	if err := r.db.Model(&models.News{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	if err := r.db.Select("id", "title", "content", "theme", "created_at").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&news).Error; err != nil {
		return nil, 0, err
	}

	return news, total, nil
}

func (r *newsRepository) GetByID(id uint) (*models.News, error) {
	var news models.News
	if err := r.db.First(&news, id).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *newsRepository) Update(news *models.News) error {
	return r.db.Save(news).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&models.News{}, id).Error
}

func (r *newsRepository) GetByTheme(theme string, page, limit int) ([]models.News, int64, error) {
	var news []models.News
	var total int64

	offset := (page - 1) * limit

	// Счетчик для темы
	if err := r.db.Model(&models.News{}).
		Where("theme = ?", theme).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Данные с фильтрацией по теме
	if err := r.db.Where("theme = ?", theme).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&news).Error; err != nil {
		return nil, 0, err
	}

	return news, total, nil
}

func (r *newsRepository) Search(query int, sortIndex string, page, limit int) ([]models.News, int64, error) {
	var news []models.News
	var total int64
	offset := (page - 1) * limit
	query_str := strconv.Itoa(query)
	var search_row = "theme = ?"
	if query == 0 {
		search_row = "theme != ?"
	}

	sort_line := "created_at DESC"
	if len(sortIndex) > 0 {
		sort_flow := " ASC"
		if strings.HasPrefix(sortIndex, "-") {
			sort_flow = " DESC"
			sortIndex = sortIndex[1:]
		}
		if sortIndex == "title" || sortIndex == "theme" || sortIndex == "created_at" || sortIndex == "id" || sortIndex == "content" {
			sort_line = sortIndex + sort_flow
		} else {
			sort_line = "created_at DESC"
		}
	}

	// Счетчик для поиска
	if err := r.db.Model(&models.News{}).
		Where(search_row, query_str).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Данные поиска
	if err := r.db.Where(search_row, query_str).
		Order(sort_line).
		Limit(limit).
		Offset(offset).
		Find(&news).Error; err != nil {
		return nil, 0, err
	}

	return news, total, nil
}
