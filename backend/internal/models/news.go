package models

import (
	"gorm.io/gorm"
	"time"
)

type NewTheme string

const (
	ThemeDev     NewTheme = "1"
	ThemeTest    NewTheme = "2"
	ThemeRelease NewTheme = "3"
	ThemeOther   NewTheme = "4"
)

type News struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title   string   `json:"title"`
	Content string   `json:"content" gorm:"type:text;not null"`
	Picture string   `json:"picture"`
	Theme   NewTheme `json:"theme"`
}
