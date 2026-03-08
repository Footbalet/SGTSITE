package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/repositories"
)

type ReleaseService interface {
	CreateRelease(version, platform, changes, checksum string, isActive bool, file *multipart.FileHeader) (*models.Release, error)
	GetAllReleases(page, limit int, platform, sortIndex string) ([]models.Release, int64, error)
	GetReleaseByID(id uint) (*models.Release, error)
	GetReleaseByVersionPlatform(version, platform string) (*models.Release, error)
	GetLatestRelease(platform string) (*models.Release, error)
	UpdateRelease(id uint, version, platform, changes, checksum string, isActive bool, file *multipart.FileHeader) (*models.Release, error)
	DeleteRelease(id uint) error
	DownloadRelease(id uint, ip string) (*models.Release, string, error)
}

type releaseService struct {
	releaseRepo repositories.ReleaseRepository
	uploadDir   string
	baseURL     string
}

func NewReleaseService(releaseRepo repositories.ReleaseRepository, uploadDir, baseURL string) ReleaseService {
	// Создаем директорию для загрузок если не существует
	os.MkdirAll(uploadDir, 0755)

	return &releaseService{
		releaseRepo: releaseRepo,
		uploadDir:   uploadDir,
		baseURL:     baseURL,
	}
}

func (s *releaseService) CreateRelease(version, platform, changes, checksum string, isActive bool, file *multipart.FileHeader) (*models.Release, error) {
	// Проверяем уникальность версии для платформы
	existing, _ := s.releaseRepo.GetByVersionPlatform(version, platform)
	if existing != nil {
		return nil, errors.New("release with this version and platform already exists")
	}

	// Сохраняем файл
	filename, fileSize, err := s.saveFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Создаем URL для доступа к файлу
	fileURL := fmt.Sprintf("%s/uploads/%s", s.baseURL, filename)

	release := &models.Release{
		Version:   version,
		Platform:  platform,
		Changes:   changes,
		FileURL:   fileURL,
		FileSize:  fileSize,
		Filename:  file.Filename,
		IsActive:  isActive,
		Checksum:  checksum,
		Downloads: 0,
	}

	if err := s.releaseRepo.Create(release); err != nil {
		// Удаляем файл если не удалось создать запись
		os.Remove(filepath.Join(s.uploadDir, filename))
		return nil, err
	}

	return release, nil
}

func (s *releaseService) GetAllReleases(page, limit int, platform, sortIndex string) ([]models.Release, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.releaseRepo.GetAll(page, limit, platform, sortIndex)
}

func (s *releaseService) GetReleaseByID(id uint) (*models.Release, error) {
	return s.releaseRepo.GetByID(id)
}

func (s *releaseService) GetReleaseByVersionPlatform(version, platform string) (*models.Release, error) {
	return s.releaseRepo.GetByVersionPlatform(version, platform)
}

func (s *releaseService) GetLatestRelease(platform string) (*models.Release, error) {
	return s.releaseRepo.GetLatest(platform)
}

func (s *releaseService) UpdateRelease(id uint, version, platform, changes, checksum string, isActive bool, file *multipart.FileHeader) (*models.Release, error) {
	// Получаем существующий релиз
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Проверяем уникальность версии если она меняется
	if version != "" && version != release.Version {
		existing, _ := s.releaseRepo.GetByVersionPlatform(version, platform)
		if existing != nil && existing.ID != id {
			return nil, errors.New("release with this version and platform already exists")
		}
		release.Version = version
	}

	if platform != "" {
		release.Platform = platform
	}
	if changes != "" {
		release.Changes = changes
	}
	if checksum != "" {
		release.Checksum = checksum
	}

	// Если загружен новый файл
	if file != nil {
		// Удаляем старый файл
		oldFilename := filepath.Base(release.FileURL)
		os.Remove(filepath.Join(s.uploadDir, oldFilename))

		// Сохраняем новый файл
		filename, fileSize, err := s.saveFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to save file: %w", err)
		}

		release.FileURL = fmt.Sprintf("%s/uploads/%s", s.baseURL, filename)
		release.FileSize = fileSize
		release.Filename = file.Filename
	}

	if err := s.releaseRepo.Update(release); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *releaseService) DeleteRelease(id uint) error {
	// Получаем релиз чтобы удалить файл
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Удаляем файл
	filename := filepath.Base(release.FileURL)
	os.Remove(filepath.Join(s.uploadDir, filename))

	// Удаляем запись из БД
	return s.releaseRepo.Delete(id)
}

type DownloadTrackerMap struct {
	downloaded map[string]bool
}

func NewDownloadTrackerMap() *DownloadTrackerMap {
	return &DownloadTrackerMap{
		downloaded: make(map[string]bool),
	}
}

func (dt *DownloadTrackerMap) AddIfNotExists(key string) bool {

	if dt.downloaded[key] {
		return false // Уже существует
	}

	dt.downloaded[key] = true
	return true // Успешно добавлен
}

var downloadedMap = NewDownloadTrackerMap()

func (s *releaseService) DownloadRelease(id uint, ip string) (*models.Release, string, error) {
	totalCode := fmt.Sprintf("%d|%s", id, ip)
	wasNotDownloaded := downloadedMap.AddIfNotExists(totalCode)
	// Получаем релиз
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return nil, "", err
	}

	// Проверяем активен ли релиз
	if !release.IsActive {
		return nil, "", errors.New("release is not active")
	}

	// Увеличиваем счетчик скачиваний
	if wasNotDownloaded {
		if err := s.releaseRepo.IncrementDownloads(id); err != nil {
			return nil, "", err
		}
	}
	// Путь к файлу
	filePath := filepath.Join(s.uploadDir, filepath.Base(release.FileURL))

	return release, filePath, nil
}

// Вспомогательные методы

func (s *releaseService) saveFile(file *multipart.FileHeader) (string, int64, error) {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return "", 0, err
	}
	defer src.Close()

	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(file.Filename, ext), ext)

	// Создаем файл на диске
	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	// Копируем содержимое
	fileSize, err := io.Copy(dst, src)
	if err != nil {
		return "", 0, err
	}

	// Проверяем размер файла (макс 500MB)
	if fileSize > 500*1024*1024 {
		os.Remove(filepath.Join(s.uploadDir, filename))
		return "", 0, errors.New("file size exceeds 500MB limit")
	}

	return filename, fileSize, nil
}
