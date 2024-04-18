package ksql

import (
	"github.com/pkg/errors"
	"github.com/thmeitz/ksqldb-go"
	"github.com/thmeitz/ksqldb-go/net"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Statistics struct {
	EventsTotal                  int               `json:"events_total"`
	EventsPerMinute              float64           `json:"events_per_minute"`
	EventsPerSecondMovingAverage float64           `json:"events_per_second_moving_average"`
	TopPlayerBets                TopPlayerBets     `json:"top_player_bets"`
	TopPlayerWins                TopPlayerWins     `json:"top_player_wins"`
	TopPlayerDeposits            TopPlayerDeposits `json:"top_player_deposits"`
}

type TopPlayerBets struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

type TopPlayerWins struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

type TopPlayerDeposits struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

type Client interface {
	GetPlayerStatistics(ctx context.Context) (*Statistics, error)
}

type ClientV1 struct {
	client *ksqldb.KsqldbClient
	logger *zap.Logger
}

type Configuration struct {
	BaseUrl  string
	Username string
	Password string
}

func NewClientV1(logger *zap.Logger, cfg Configuration) *ClientV1 {
	options := net.Options{
		Credentials: net.Credentials{Username: cfg.Username, Password: cfg.Password},
		BaseUrl:     cfg.BaseUrl,
		AllowHTTP:   true,
	}

	kcl, err := ksqldb.NewClientWithOptions(options)
	if err != nil {
		logger.Fatal("Failed to create KSQL client", zap.Error(err))
	}
	defer kcl.Close()

	return &ClientV1{
		logger: logger.Named("ksql-client-v1"),
		client: &kcl,
	}
}

func (c *ClientV1) GetPlayerStatistics(ctx context.Context) (*Statistics, error) {
	query := `select timestamptostring(windowstart,'yyyy-MM-dd HH:mm:ss','Europe/London') as window_start,
timestamptostring(windowend,'HH:mm:ss','Europe/London') as window_end, dog_size, dogs_ct
from dogs_by_size where dog_size=?;`

	stmnt, err := ksqldb.QueryBuilder(query, "dogsize")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	qOpts := (&ksqldb.QueryOptions{Sql: *stmnt}).EnablePullQueryTableScan(false)

	_, row, err := c.client.Pull(ctx, *qOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	return nil, nil
}
