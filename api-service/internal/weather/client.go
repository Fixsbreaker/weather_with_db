package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type WeatherData struct {
	Temperature float64
	Description string
}

// wttr.in JSON response shapes (only fields we need)
type wttrResponse struct {
	CurrentCondition []struct {
		TempC       string `json:"temp_C"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
}

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 5 * time.Second},
		baseURL: "https://wttr.in",
	}
}

func (c *Client) GetWeather(ctx context.Context, city string) (*WeatherData, error) {
	// url.PathEscape handles spaces and special chars in city names safely
	endpoint := fmt.Sprintf("%s/%s?format=j1", c.baseURL, url.PathEscape(city))

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

	var wttr wttrResponse
	if err := json.NewDecoder(resp.Body).Decode(&wttr); err != nil {
		return nil, fmt.Errorf("decode weather response: %w", err)
	}

	if len(wttr.CurrentCondition) == 0 {
		return nil, fmt.Errorf("no weather data for city %q", city)
	}

	cond := wttr.CurrentCondition[0]
	temp, err := strconv.ParseFloat(cond.TempC, 64)
	if err != nil {
		return nil, fmt.Errorf("parse temperature: %w", err)
	}

	desc := ""
	if len(cond.WeatherDesc) > 0 {
		desc = cond.WeatherDesc[0].Value
	}

	return &WeatherData{
		Temperature: temp,
		Description: desc,
	}, nil
}
