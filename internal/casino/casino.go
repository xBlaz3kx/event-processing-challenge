package casino

import (
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/ksql"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Service interface {
	GetStatistics(ctx context.Context) (*Statistics, error)
}

type ServiceV1 struct {
	client ksql.Client
	logger *zap.Logger
}

func NewServiceV1(logger *zap.Logger, client ksql.Client) *ServiceV1 {
	return &ServiceV1{
		logger: logger.Named("casino-service"),
		client: client,
	}
}

func (s ServiceV1) GetStatistics(ctx context.Context) (*Statistics, error) {
	s.logger.Info("Getting statistics")

	_, err := s.client.GetPlayerStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
	// return statistics.(*Statistics), nil
}
