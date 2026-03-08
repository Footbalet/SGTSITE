package services

import (
	"backend/internal/models"
	"backend/internal/repositories"
	"errors"
)

type NewsService interface {
	CreateNews(news *models.News) error
	GetAllNews(page, limit int) ([]models.News, int64, error)
	GetNewsByID(id uint) (*models.News, error)
	UpdateNews(id uint, updatedNews *models.News) error
	DeleteNews(id uint) error
	GetNewsByTheme(theme string, page, limit int) ([]models.News, int64, error)
	SearchNews(query int, sortIndex string, page, limit int) ([]models.News, int64, error)
}

type newsService struct {
	newsRepo repositories.NewsRepository
}

func NewNewsService(newsRepo repositories.NewsRepository) NewsService {
	return &newsService{newsRepo: newsRepo}
}

func (s *newsService) CreateNews(news *models.News) error {
	// Валидация
	if news.Title == "" {
		return errors.New("title is required")
	}
	if news.Content == "" {
		return errors.New("content is required")
	}

	return s.newsRepo.Create(news)
}

func (s *newsService) GetAllNews(page, limit int) ([]models.News, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return s.newsRepo.GetAll(page, limit)
}

func (s *newsService) GetNewsByID(id uint) (*models.News, error) {
	return s.newsRepo.GetByID(id)
}

func (s *newsService) UpdateNews(id uint, updatedNews *models.News) error {
	// Получаем существующую новость
	existingNews, err := s.newsRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Обновляем поля
	if updatedNews.Title != "" {
		existingNews.Title = updatedNews.Title
	}
	if updatedNews.Content != "" {
		existingNews.Content = updatedNews.Content
	}
	if updatedNews.Picture != "" {
		existingNews.Picture = updatedNews.Picture
	}
	if updatedNews.Theme != "" {
		existingNews.Theme = updatedNews.Theme
	}

	return s.newsRepo.Update(existingNews)
}

func (s *newsService) DeleteNews(id uint) error {
	// Проверяем существование
	_, err := s.newsRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.newsRepo.Delete(id)
}

func (s *newsService) GetNewsByTheme(theme string, page, limit int) ([]models.News, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return s.newsRepo.GetByTheme(theme, page, limit)
}

func (s *newsService) SearchNews(query int, sortIndex string, page, limit int) ([]models.News, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return s.newsRepo.Search(query, sortIndex, page, limit)
}
