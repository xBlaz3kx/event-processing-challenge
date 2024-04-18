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
	query := `select timestamptostring(windowstart,'yyyy-MM-dd HH:mm:ss','Europe/London') as window_start,
			timestamptostring(windowend,'HH:mm:ss','Europe/London') as window_end, dog_size, dogs_ct
				from dogs_by_size where dog_size=?;`

	statement, err := ksqldb.QueryBuilder(query, "dogsize")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	queryOpts := &ksqldb.QueryOptions{Sql: *statement}
	queryOpts.EnablePullQueryTableScan(false)

	_, rows, err := c.client.Pull(ctx, *queryOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	response := &Statistics{}
	for _, row := range rows {
		if row != nil {
			// Should do some type assertions here
			response.EventsTotal = row[0].(int)
			response.EventsPerMinute = row[1].(float64)
			response.EventsPerSecondMovingAverage = row[2].(float64)
			response.TopPlayerBets.Id = row[3].(int)
			response.TopPlayerBets.Count = row[4].(int)
			response.TopPlayerWins.Id = row[5].(int)
			response.TopPlayerWins.Count = row[6].(int)
			response.TopPlayerDeposits.Id = row[7].(int)
			response.TopPlayerDeposits.Count = row[8].(int)
			break
		}
	}

	return response, nil
}

func (c *KsqlClientV1) Close() {
	c.client.Close()
}
