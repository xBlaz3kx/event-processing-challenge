package currency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/cache"
	"go.uber.org/zap"
)

type Service interface {
	Convert(ctx context.Context, from, to string, amount int) (int, error)
}

type ServiceV1 struct {
	currencyCache   cache.Cache
	exchangeRateApi *exchangeRateApiClient
	logger          *zap.Logger
}

func NewServiceV1(logger *zap.Logger, currencyCache cache.Cache, config ExchangeApiConfig) *ServiceV1 {
	return &ServiceV1{
		currencyCache:   currencyCache,
		exchangeRateApi: newExchangeRateApiClient(logger, config),
		logger:          logger.Named("currency-service-v1"),
	}
}

// Convert converts the amount from one currency to another. If the cache is enabled, it will first check the cache for the conversion rate.
func (s *ServiceV1) Convert(ctx context.Context, from, to string, amount int) (int, error) {
	logger := s.logger.With(zap.String("from", from), zap.String("to", to), zap.Int("amount", amount))
	logger.Info("Converting currency")

	// Check if from currency is equal to currency
	if from == to {
		logger.Debug("from currency is equal to currency")
		return amount, nil
	}

	// Check if the conversion is cached
	cacheKey := fmt.Sprintf("%s-%s", from, to)
	if rate, err := s.currencyCache.Get(ctx, cacheKey); err == nil {
		logger.Debug("Cache hit, using the exchange rate from cache", zap.Float64("rate", *rate))
		return int(ConvertToStandardUnit(amount, from) * *rate), nil
	}

	// If not cached, fetch the conversion rate from the API and cache it
	rate, err := s.exchangeRateApi.getExchangeRate(ctx, from, to)
	if err != nil {
		logger.Error("failed to get exchange rate", zap.Error(err))
		return 0, err
	}

	// Cache the exchange rate
	if err := s.currencyCache.Set(ctx, cacheKey, *rate); err != nil {
		logger.Warn("failed to cache exchange rate", zap.Error(err))
	}

	return int(ConvertToStandardUnit(amount, from) * *rate), nil
}

// ConvertToStandardUnit converts the amount from the smallest unit to the standard unit of the given currency.
func ConvertToStandardUnit(amount int, currency string) float64 {
	switch strings.ToUpper(currency) {
	case "EUR", "USD", "GBP", "NZD":
		return float64(amount) / 100
	case "BTC":
		return float64(amount) / 100000000
	default:
		return float64(amount)
	}
}

func (s *ServiceV1) Pass() bool {
	ctx, end := context.WithTimeout(context.Background(), 5*time.Second)
	defer end()

	// Ping the api if its reachable
	return s.exchangeRateApi.ping(ctx)
}

func (s *ServiceV1) Name() string {
	return "currency-service-v1"
}
