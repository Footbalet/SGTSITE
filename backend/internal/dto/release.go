package dto

import (
	"mime/multipart"
	"time"
)

type CreateReleaseRequest struct {
	Version  string                `form:"version" binding:"required,min=1,max=50"`
	Platform string                `form:"platform" binding:"required,oneof=Windows Linux Mac Android IOS Web"`
	Changes  string                `form:"changes" binding:"required,min=10"`
	IsActive *bool                 `form:"is_active"`
	File     *multipart.FileHeader `form:"file" binding:"required"`
}

type UpdateReleaseRequest struct {
	Version  string                `form:"version" binding:"omitempty,min=1,max=50"`
	Platform string                `form:"platform" binding:"omitempty,oneof=Windows Linux Mac Android IOS Web"`
	Changes  string                `form:"changes" binding:"omitempty,min=10"`
	File     *multipart.FileHeader `form:"file"`
}

type ReleaseResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version"`
	Platform  string    `json:"platform"`
	Changes   string    `json:"changes"`
	FileURL   string    `json:"file_url"`
	Downloads uint      `json:"downloads"`
}
