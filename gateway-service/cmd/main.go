package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
}

type wttrResponse struct {
	CurrentCondition []struct {
		TempC       string `json:"temp_C"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			http.Error(w, "city is required", http.StatusBadRequest)
			return
		}

		endpoint := fmt.Sprintf("https://wttr.in/%s?format=j1", url.PathEscape(city))
		
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("fetch weather error: %v", err)
			http.Error(w, "failed to fetch weather", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("weather API returned status %d", resp.StatusCode)
			http.Error(w, "weather API error", http.StatusBadGateway)
			return
		}

		var wttr wttrResponse
		if err := json.NewDecoder(resp.Body).Decode(&wttr); err != nil {
			log.Printf("decode error: %v", err)
			http.Error(w, "failed to decode response", http.StatusInternalServerError)
			return
		}

		if len(wttr.CurrentCondition) == 0 {
			http.Error(w, "no weather data", http.StatusNotFound)
			return
		}

		cond := wttr.CurrentCondition[0]
		temp, err := strconv.ParseFloat(cond.TempC, 64)
		if err != nil {
			log.Printf("parse temperature error: %v", err)
			http.Error(w, "invalid temperature data", http.StatusInternalServerError)
			return
		}

		desc := ""
		if len(cond.WeatherDesc) > 0 {
			desc = cond.WeatherDesc[0].Value
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(WeatherData{
			Temperature: temp,
			Description: desc,
		})
	})

	log.Printf("Gateway service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
