package kafka

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

type KsqlClient interface {
	GetPlayerStatistics(ctx context.Context) (*Statistics, error)
}

type KsqlClientV1 struct {
	client *ksqldb.KsqldbClient
	logger *zap.Logger
}

type KsqlConfiguration struct {
	BaseUrl  string `yaml:"baseUrl"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func NewClientV1(logger *zap.Logger, cfg KsqlConfiguration) *KsqlClientV1 {
	options := net.Options{
		Credentials: net.Credentials{Username: cfg.Username, Password: cfg.Password},
		BaseUrl:     cfg.BaseUrl,
		AllowHTTP:   true,
	}

	kcl, err := ksqldb.NewClientWithOptions(options)
	if err != nil {
		logger.Fatal("Failed to create KSQL client", zap.Error(err))
	}

	return &KsqlClientV1{
		logger: logger.Named("ksql-client-v1"),
		client: &kcl,
	}
}

func (c *KsqlClientV1) GetPlayerStatistics(ctx context.Context) (*Statistics, error) {
	response := &Statistics{}

	total, perMin, err := c.getEventsTotal(ctx)
	if err != nil {
		return nil, err
	}

	response.EventsTotal = *total
	response.EventsPerMinute = *perMin

	return response, nil
}

func (c *KsqlClientV1) getEventsTotal(ctx context.Context) (*int, *float64, error) {
	query := `SELECT TOTAL_EVENTS FROM CASINO_EVENTS_STATS;`

	statement, err := ksqldb.QueryBuilder(query)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build query")
	}

	queryOpts := ksqldb.QueryOptions{Sql: *statement}

	_, rows, err := c.client.Pull(ctx, queryOpts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to execute query")
	}

	for _, row := range rows {
		if row != nil {
			totalEvents := row[0].(int)
			eventsPerMin := float64(totalEvents / 60.0)
			return &totalEvents, &eventsPerMin, nil
		}
	}

	return nil, nil, errors.New("failed to get events total")
}

func (c *KsqlClientV1) getTopPlayers(ctx context.Context) (*int, *float64, error) {
	query := `SELECT player_id, player_bets, player_wins, player_deposits FROM CASINO_PLAYER_STATS;`

	statement, err := ksqldb.QueryBuilder(query)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build query")
	}

	queryOpts := ksqldb.QueryOptions{Sql: *statement}

	_, rows, err := c.client.Pull(ctx, queryOpts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to execute query")
	}

	for _, row := range rows {
		if row != nil {
			totalEvents := row[0].(int)
			eventsPerMin := float64(totalEvents / 60.0)
			return &totalEvents, &eventsPerMin, nil
		}
	}

	return nil, nil, errors.New("failed to get events total")
}

func (c *KsqlClientV1) Close() {
	c.client.Close()
}
