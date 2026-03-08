package models

import (
	"gorm.io/gorm"
	"time"
)

type Release struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Version   string `json:"version" gorm:"not null;uniqueIndex:idx_version_platform"`  // Уникальная версия для платформы
	Platform  string `json:"platform" gorm:"not null;uniqueIndex:idx_version_platform"` // windows, linux, mac, android, ios
	Changes   string `json:"changes" gorm:"type:text;not null"`
	FileURL   string `json:"file_url" gorm:"not null"` // URL или путь к файлу
	FileSize  int64  `json:"file_size"`                // Размер файла в байтах
	Filename  string `json:"filename"`                 // Оригинальное имя файла
	Downloads uint   `json:"downloads" gorm:"default:0"`
	IsActive  bool   `json:"is_active" gorm:"default:true"` // Активен ли релиз для скачивания
	Checksum  string `json:"checksum"`                      // MD5/SHA256 для проверки целостности
}
