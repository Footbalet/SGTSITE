package repositories

import (
	"backend/internal/models"
	"gorm.io/gorm"
	"strings"
)

type ReleaseRepository interface {
	Create(release *models.Release) error
	GetAll(page, limit int, platform, sortIndex string) ([]models.Release, int64, error)
	GetByID(id uint) (*models.Release, error)
	GetByVersionPlatform(version, platform string) (*models.Release, error)
	GetLatest(platform string) (*models.Release, error)
	Update(release *models.Release) error
	Delete(id uint) error
	IncrementDownloads(id uint) error
}

type releaseRepository struct {
	db *gorm.DB
}

func NewReleaseRepository(db *gorm.DB) ReleaseRepository {
	return &releaseRepository{db: db}
}

func (r *releaseRepository) Create(release *models.Release) error {
	return r.db.Create(release).Error
}

func (r *releaseRepository) GetAll(page, limit int, platform, sortIndex string) ([]models.Release, int64, error) {
	var releases []models.Release
	var total int64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Release{})

	// Фильтр по платформе если указан
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	sort_line := "created_at DESC"
	if len(sortIndex) > 0 {
		sort_flow := " ASC"
		if strings.HasPrefix(sortIndex, "-") {
			sort_flow = " DESC"
			sortIndex = sortIndex[1:]
		}
		if sortIndex == "platform" || sortIndex == "created_at" || sortIndex == "downloads" || sortIndex == "id" || sortIndex == "version" {
			sort_line = sortIndex + sort_flow
		} else {
			sort_line = "created_at DESC"
		}
	}
	// Получаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	if err := query.Select("id", "version", "changes", "platform", "downloads", "file_url", "created_at").
		Order(sort_line).
		Limit(limit).
		Offset(offset).
		Find(&releases).Error; err != nil {
		return nil, 0, err
	}

	return releases, total, nil
}

func (r *releaseRepository) GetByID(id uint) (*models.Release, error) {
	var release models.Release
	if err := r.db.First(&release, id).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *releaseRepository) GetByVersionPlatform(version, platform string) (*models.Release, error) {
	var release models.Release
	if err := r.db.Where("version = ? AND platform = ?", version, platform).
		First(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *releaseRepository) GetLatest(platform string) (*models.Release, error) {
	var release models.Release
	query := r.db.Where("is_active = true")

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if err := query.Order("created_at DESC").First(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *releaseRepository) Update(release *models.Release) error {
	return r.db.Save(release).Error
}

func (r *releaseRepository) Delete(id uint) error {
	return r.db.Delete(&models.Release{}, id).Error
}

func (r *releaseRepository) IncrementDownloads(id uint) error {
	return r.db.Model(&models.Release{}).
		Where("id = ?", id).
		Update("downloads", gorm.Expr("downloads + ?", 1)).Error
}
