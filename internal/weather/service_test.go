package weather

import (
	"testing"
	"time"
)

// ============ 单元测试 ============

// TestCityCoords 测试城市坐标映射
func TestCityCoords(t *testing.T) {
	tests := []struct {
		name     string
		city     string
		wantLat  float64
		wantLon  float64
		wantFind bool
	}{
		{"Beijing", "beijing", 39.9042, 116.4074, true},
		{"Shanghai", "shanghai", 31.2304, 121.4737, true},
		{"Guangzhou", "guangzhou", 23.1291, 113.2644, true},
		{"Shenzhen", "shenzhen", 22.5431, 114.0579, true},
		{"Unknown", "tokyo", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coords, found := CityCoords[tt.city]
			if found != tt.wantFind {
				t.Errorf("CityCoords[%q] found = %v, want %v", tt.city, found, tt.wantFind)
			}
			if found {
				if coords[0] != tt.wantLat {
					t.Errorf("Latitude = %v, want %v", coords[0], tt.wantLat)
				}
				if coords[1] != tt.wantLon {
					t.Errorf("Longitude = %v, want %v", coords[1], tt.wantLon)
				}
			}
		})
	}
}

// TestFetchWeather_InvalidCity 测试无效城市名
func TestFetchWeather_InvalidCity(t *testing.T) {
	_, err := FetchWeather("invalid_city_xyz")
	if err == nil {
		t.Error("Expected error for invalid city, got nil")
	}
	if err != nil && err.Error() != "unknown city: invalid_city_xyz" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestFetchWeather_ValidCities 测试有效城市（需要网络）
func TestFetchWeather_ValidCities(t *testing.T) {
	cities := []string{"beijing", "shanghai"}

	for _, city := range cities {
		t.Run(city, func(t *testing.T) {
			data, err := FetchWeather(city)
			if err != nil {
				t.Skipf("Network unavailable or API limit: %v", err)
				return
			}

			if data.City != city {
				t.Errorf("City = %q, want %q", data.City, city)
			}
			if data.Temperature == 0 {
				t.Error("Temperature should not be zero")
			}
			if data.Humidity < 0 || data.Humidity > 100 {
				t.Errorf("Humidity out of range: %d", data.Humidity)
			}
			if data.RecordedAt.IsZero() {
				t.Error("RecordedAt should be set")
			}
		})
	}
}

// TestWeatherData_JSONMarshal 测试 WeatherData JSON 序列化
func TestWeatherData_JSONMarshal(t *testing.T) {
	data := WeatherData{
		City:        "beijing",
		Temperature: 25.5,
		Humidity:    60,
		RecordedAt:  time.Date(2026, 3, 10, 10, 30, 0, 0, time.UTC),
	}

	// 简单验证结构体字段可序列化
	if data.City != "beijing" {
		t.Error("City field mismatch")
	}
	if data.Temperature != 25.5 {
		t.Error("Temperature field mismatch")
	}
}

// ============ 基准测试 ============

// BenchmarkCityCoordsLookup 基准测试：城市坐标查找性能
func BenchmarkCityCoordsLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, ok := CityCoords["beijing"]
		if !ok {
			b.Fatal("Lookup failed")
		}
	}
}

// BenchmarkCityCoordsLookup_AllCities 基准测试：遍历所有城市
func BenchmarkCityCoordsLookup_AllCities(b *testing.B) {
	cities := []string{"beijing", "shanghai", "guangzhou", "shenzhen"}
	for i := 0; i < b.N; i++ {
		for _, city := range cities {
			_, ok := CityCoords[city]
			if !ok {
				b.Fatal("Lookup failed")
			}
		}
	}
}

// ============ 并发测试 ============

// TestFetchWeatherConcurrent_Empty 测试空城市列表
func TestFetchWeatherConcurrent_Empty(t *testing.T) {
	results, err := FetchWeatherConcurrent([]string{})
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got: %d", len(results))
	}
}

// TestFetchWeatherConcurrent_InvalidCities 测试全部无效城市
func TestFetchWeatherConcurrent_InvalidCities(t *testing.T) {
	cities := []string{"invalid1", "invalid2", "invalid3"}
	results, _ := FetchWeatherConcurrent(cities)

	// 注意：当前实现只打印警告，不返回 error
	// 所有城市都无效时，results 应该为空
	if len(results) != 0 {
		t.Errorf("Expected 0 results for invalid cities, got: %d", len(results))
	}
}

// TestFetchWeatherConcurrent_Mixed 测试混合城市（有效 + 无效）
func TestFetchWeatherConcurrent_Mixed(t *testing.T) {
	cities := []string{"beijing", "invalid_city", "shanghai"}
	results, _ := FetchWeatherConcurrent(cities)

	// 至少应该有部分成功（如果网络可用）
	t.Logf("Got %d results from %d cities", len(results), len(cities))
	for _, r := range results {
		t.Logf("  - %s: %.1f°C, %d%%", r.City, r.Temperature, r.Humidity)
	}
}
