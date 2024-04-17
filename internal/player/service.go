package player

import (
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"github.com/xBlaz3kx/event-processing-challenge/internal/player/database"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Service interface {
	GetPlayerDetails(ctx context.Context, playerID int) (*casino.Player, error)
	GetStatistics(ctx context.Context) (*Statistics, error)
}

type ServiceV1 struct {
	logger     *zap.Logger
	repository database.Repository
}

func NewServiceV1(logger *zap.Logger, repository database.Repository) *ServiceV1 {
	return &ServiceV1{
		logger:     logger.Named("player-service-v1"),
		repository: repository,
	}
}

// GetPlayerDetails returns player details from the database.
func (s *ServiceV1) GetPlayerDetails(ctx context.Context, playerID int) (*casino.Player, error) {
	logger := s.logger.With(zap.Int("playerID", playerID))
	logger.Info("Getting player details")

	details, err := s.repository.GetPlayerDetails(ctx, playerID)
	if err != nil {
		logger.Error("Failed to get player details", zap.Error(err))
		return nil, err
	}

	return details, nil
}

func (s *ServiceV1) GetStatistics(ctx context.Context) (*Statistics, error) {
	s.logger.Info("Getting statistics")

	//TODO implement me
	panic("implement me")
}

// Pass checks if the service is healthy.
func (s *ServiceV1) Pass() bool {
	return true
}

// Name returns the name of the service.
func (s *ServiceV1) Name() string {
	return "player-service-v1"
}
