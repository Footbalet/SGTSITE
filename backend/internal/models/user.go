package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `json:"-"` // Пароль не возвращается в JSON
	IsActive bool   `gorm:"default:true" json:"is_active"`
	IsAdmin  bool   `gorm:"default:false" json:"is_admin"`

	// Дополнительные поля
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
