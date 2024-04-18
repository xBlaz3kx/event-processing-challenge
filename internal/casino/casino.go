package casino

import (
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Service interface {
	GetStatistics(ctx context.Context) (*kafka.Statistics, error)
}

type ServiceV1 struct {
	client kafka.KsqlClient
	logger *zap.Logger
}

func NewServiceV1(logger *zap.Logger, client kafka.KsqlClient) *ServiceV1 {
	return &ServiceV1{
		logger: logger.Named("casino-service"),
		client: client,
	}
}

func (s ServiceV1) GetStatistics(ctx context.Context) (*kafka.Statistics, error) {
	s.logger.Info("Getting statistics")

	stats, err := s.client.GetPlayerStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
