package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
}

type Client struct {
	http       *http.Client
	gatewayURL string
}

func NewClient(gatewayURL string) *Client {
	return &Client{
		http:       &http.Client{Timeout: 5 * time.Second},
		gatewayURL: gatewayURL,
	}
}

func (c *Client) GetWeather(ctx context.Context, city string) (*WeatherData, error) {
	endpoint := fmt.Sprintf("%s/weather?city=%s", c.gatewayURL, url.QueryEscape(city))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch weather for %q: %w", city, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d for city %q", resp.StatusCode, city)
	}

	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode weather response: %w", err)
	}

	return &data, nil
}
