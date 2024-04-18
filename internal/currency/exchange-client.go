package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type ExchangeApiConfig struct {
	Url    string `yaml:"url"`
	ApiKey string `yaml:"apiKey"`
}

type latestResponse struct {
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
}

type exchangeRateApiClient struct {
	*http.Client
	logger *zap.Logger
	config ExchangeApiConfig
}

func newExchangeRateApiClient(logger *zap.Logger, config ExchangeApiConfig) *exchangeRateApiClient {
	return &exchangeRateApiClient{
		Client: &http.Client{},
		config: config,
		logger: logger,
	}
}

// getExchangeRate fetches the exchange rate from the API.
func (c *exchangeRateApiClient) getExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*float64, error) {
	c.logger.Debug("Fetching exchange rate from API", zap.String("from", fromCurrency), zap.String("to", toCurrency))

	url := fmt.Sprintf("%s/latest?symbols=%s&base=%s", c.config.Url, toCurrency, fromCurrency)
	c.logger.Debug("URL", zap.String("url", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create a request")
	}

	req.Header.Set("apiKey", c.config.ApiKey)
	req.WithContext(ctx)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot execute a request")
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response
	response := latestResponse{}
	err = json.Unmarshal(all, &response)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal response")
	}

	if !response.Success {
		return nil, errors.New("unable to fetch exchange rate")
	}

	rate, found := response.Rates[toCurrency]
	if !found {
		return nil, errors.New("currency not found")
	}

	return &rate, nil
}

// ping checks if the API is reachable.
func (c *exchangeRateApiClient) ping(ctx context.Context) bool {
	req, err := http.NewRequest("GET", c.config.Url, nil)
	if err != nil {
		c.logger.Error("Cannot create a request", zap.Error(err))
		return false
	}

	req.WithContext(ctx)
	req.Header.Set("apiKey", c.config.ApiKey)
	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Error("Unable to perform request", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	// Success if the status code is between 200 and 299
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}

	return false
}
