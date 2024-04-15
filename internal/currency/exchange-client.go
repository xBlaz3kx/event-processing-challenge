package currency

import (
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type exchangeRateApiClient struct {
	*http.Client
	url    string
	logger *zap.Logger
}

func newExchangeRateApiClient(logger *zap.Logger) *exchangeRateApiClient {
	return &exchangeRateApiClient{
		Client: &http.Client{},
		url:    "https://api.exchangeratesapi.io/latest",
		logger: logger,
	}
}

// getExchangeRate fetches the exchange rate from the API.
func (c *exchangeRateApiClient) getExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*float64, error) {

	return nil, nil
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
