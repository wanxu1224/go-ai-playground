package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenMeteoResponse API 响应结构
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Current   struct {
		Temperature2M float64 `json:"temperature_2m"`
		Humidity      int     `json:"relative_humidity_2m"`
		Time          string  `json:"time"`
	} `json:"current"`
}

// WeatherData 标准化天气数据
type WeatherData struct {
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	RecordedAt  time.Time `json:"recorded_at"`
}

// CityCoords 城市坐标映射
var CityCoords = map[string][2]float64{
	"beijing":  {39.9042, 116.4074},
	"shanghai": {31.2304, 121.4737},
	"guangzhou": {23.1291, 113.2644},
	"shenzhen": {22.5431, 114.0579},
}

// FetchWeather 获取单个城市天气
func FetchWeather(city string) (*WeatherData, error) {
	coords, ok := CityCoords[city]
	if !ok {
		return nil, fmt.Errorf("unknown city: %s", city)
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m",
		coords[0], coords[1],
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp OpenMeteoResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &WeatherData{
		City:        city,
		Temperature: apiResp.Current.Temperature2M,
		Humidity:    apiResp.Current.Humidity,
		RecordedAt:  time.Now(),
	}, nil
}

// FetchWeatherConcurrent 并发获取多个城市天气
func FetchWeatherConcurrent(cities []string) ([]WeatherData, error) {
	results := make([]WeatherData, 0, len(cities))
	errors := make([]error, 0)

	done := make(chan *WeatherData)
	errChan := make(chan error)

	// 启动 goroutine 并发请求
	for _, city := range cities {
		go func(c string) {
			data, err := FetchWeather(c)
			if err != nil {
				errChan <- fmt.Errorf("%s: %v", c, err)
				return
			}
			done <- data
		}(city)
	}

	// 收集结果
	for i := 0; i < len(cities); i++ {
		select {
		case data := <-done:
			results = append(results, *data)
		case err := <-errChan:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("⚠️ Partial errors: %v\n", errors)
	}

	return results, nil
}
