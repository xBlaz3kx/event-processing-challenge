package cache

import (
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type RedisConfiguration struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RedisCache struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewRedisCache(cfg RedisConfiguration, logger *zap.Logger) *RedisCache {
	return &RedisCache{
		logger: logger,
		redisClient: redis.NewClient(&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

// Get retrieves a value from the cache.
func (r *RedisCache) Get(ctx context.Context, key string) (*float64, error) {
	stat := r.redisClient.Get(ctx, key)
	if stat == nil {
		return nil, errors.New("error getting value from redis cache")
	}

	if stat.Err() != nil {
		return nil, errors.Wrap(stat.Err(), "error getting value from redis cache")
	}

	return nil, nil
}

// Set sets a value in the cache. The value will expire after 1 minute.
func (r *RedisCache) Set(ctx context.Context, key string, value float64) error {
	stat := r.redisClient.Set(ctx, key, value, time.Minute)
	if stat == nil {
		return errors.New("error setting value in redis cache")
	}

	if stat.Err() != nil {
		return errors.Wrap(stat.Err(), "error setting value in redis cache")
	}

	_, err := stat.Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisCache) Pass() bool {
	// Ping the redis server
	ctx, end := context.WithTimeout(context.Background(), 5*time.Second)
	defer end()

	cmd := r.redisClient.Ping(ctx)
	return cmd.Err() == nil
}

func (r *RedisCache) Name() string {
	return "redis-cache"
}
