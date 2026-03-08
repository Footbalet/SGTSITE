package services

import (
	"errors"
	"time"

	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/repositories"
)

type GameService interface {
	CreateGame(req dto.CreateGameRequest) (*models.Game, error)
	GetAllGames(filter repositories.GameFilter) ([]models.Game, int64, error)
	GetGameByID(id uint) (*models.Game, error)
	UpdateGame(id uint, req dto.UpdateGameRequest) (*models.Game, error)
	DeleteGame(id uint) error
	GetActiveGames(page, limit int) ([]models.Game, error)
	FinishGame(id uint) (*models.Game, error)
	CancelGame(id uint) (*models.Game, error)
	UpdatePlayerCount(id uint, currentPlayers int) (*models.Game, error)
	GetGamesByDateRange(from, to time.Time) ([]models.Game, error)
}

type gameService struct {
	gameRepo repositories.GameRepository
}

func NewGameService(gameRepo repositories.GameRepository) GameService {
	return &gameService{gameRepo: gameRepo}
}

func (s *gameService) CreateGame(req dto.CreateGameRequest) (*models.Game, error) {
	// Создаем игру
	game := &models.Game{
		MapID:          req.MapID,
		GameMode:       req.GameMode,
		Status:         models.StatusActive,
		CurrentPlayers: req.CurrentPlayers,
		MaxPlayers:     req.CurrentPlayers,
	}

	if req.Status != "" {
		game.Status = req.Status
	}

	game.CalculateDuration()

	if err := s.gameRepo.Create(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) GetAllGames(filter repositories.GameFilter) ([]models.Game, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	return s.gameRepo.GetAll(repositories.GameFilter{
		Status:   filter.Status,
		GameMode: filter.GameMode,
		MapID:    filter.MapID,
		FromDate: filter.FromDate,
		ToDate:   filter.ToDate,
		Page:     filter.Page,
		Limit:    filter.Limit,
	})
}

func (s *gameService) GetGameByID(id uint) (*models.Game, error) {
	return s.gameRepo.GetByID(id)
}

func (s *gameService) UpdateGame(id uint, req dto.UpdateGameRequest) (*models.Game, error) {
	// Получаем существующую игру
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if req.Finish != nil {
		game.Finish = *req.Finish
	}

	if req.MapID != "" {
		game.MapID = req.MapID
	}

	if req.GameMode != "" {
		game.GameMode = req.GameMode
	}

	if req.Status != "" {
		game.Status = req.Status
	}

	if req.ServerIP != "" {
		game.ServerIP = req.ServerIP
	}

	if req.MaxPlayers != nil {
		game.MaxPlayers = *req.MaxPlayers
	}

	if req.CurrentPlayers != nil {
		if *req.CurrentPlayers > game.MaxPlayers {
			return nil, errors.New("current players cannot exceed max players")
		}
		game.CurrentPlayers = *req.CurrentPlayers
	}

	// Пересчитываем длительность
	game.CalculateDuration()

	if err := s.gameRepo.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) DeleteGame(id uint) error {
	return s.gameRepo.Delete(id)
}

func (s *gameService) GetActiveGames(page, limit int) ([]models.Game, error) {
	return s.gameRepo.GetActive(page, limit)
}

func (s *gameService) FinishGame(id uint) (*models.Game, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if game.Status != models.StatusActive {
		return nil, errors.New("only active games can be finished")
	}

	game.Status = models.StatusFinished
	game.Finish = time.Now()
	game.CalculateDuration()

	if err := s.gameRepo.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) CancelGame(id uint) (*models.Game, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if game.Status == models.StatusFinished {
		return nil, errors.New("cannot cancel finished game")
	}

	game.Status = models.StatusCancelled
	if err := s.gameRepo.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) UpdatePlayerCount(id uint, currentPlayers int) (*models.Game, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	game.CurrentPlayers = currentPlayers
	if game.MaxPlayers < currentPlayers {
		game.MaxPlayers = currentPlayers
	}
	if err := s.gameRepo.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) GetGamesByDateRange(from, to time.Time) ([]models.Game, error) {
	return s.gameRepo.GetByDateRange(from, to)
}
