package currency

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type ExchangeApiConfig struct {
	Url    string `yaml:"url"`
	ApiKey string `yaml:"apiKey"`
}

type conversionResponse struct {
	Date       string `json:"date"`
	Historical string `json:"historical"`
	Info       struct {
		Rate      float64 `json:"rate"`
		Timestamp int     `json:"timestamp"`
	} `json:"info"`
	Query struct {
		Amount int    `json:"amount"`
		From   string `json:"from"`
		To     string `json:"to"`
	} `json:"query"`
	Result  float64 `json:"result"`
	Success bool    `json:"success"`
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
	url    string
	logger *zap.Logger
}

func newExchangeRateApiClient(logger *zap.Logger, config ExchangeApiConfig) *exchangeRateApiClient {
	return &exchangeRateApiClient{
		Client: &http.Client{},
		url:    config.Url,
		logger: logger,
	}
}

// getExchangeRate fetches the exchange rate from the API.
func (c *exchangeRateApiClient) getExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*float64, error) {
	c.logger.Debug("Fetching exchange rate from API", zap.String("from", fromCurrency), zap.String("to", toCurrency))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?symbols=%s&base=%s", c.url, toCurrency, fromCurrency), nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create a request")
	}

	req.WithContext(ctx)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot execute a request")
	}
	defer resp.Body.Close()

	if resp.ContentLength == 0 {
		return nil, errors.New("empty response")
	}

	bytes := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read response")
	}

	// Unmarshal the response
	response := latestResponse{}
	err = json.Unmarshal(bytes, &response)
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
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		c.logger.Error("Cannot create a request", zap.Error(err))
		return false
	}

	req.WithContext(ctx)
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
