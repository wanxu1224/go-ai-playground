package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-ai-playground/internal/db"
)

var weatherDB *db.Database

// WeatherRequest 请求结构
type WeatherRequest struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
}

// WeatherResponse 响应结构
type WeatherResponse struct {
	ID         int64   `json:"id"`
	Message    string  `json:"message"`
	City       string  `json:"city,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Humidity   int     `json:"humidity,omitempty"`
}

// HistoryResponse 历史记录响应
type HistoryResponse struct {
	Records []db.WeatherRecord `json:"records"`
}

func init() {
	// 初始化 SQLite 数据库
	var err error
	weatherDB, err = db.New("weather.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("✅ Database initialized successfully")
}

// healthHandler 健康检查
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// weatherHandler 保存天气数据
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WeatherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, err := weatherDB.SaveWeather(req.City, req.Temperature, req.Humidity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WeatherResponse{
		ID:          id,
		Message:     "Weather data saved successfully",
		City:        req.City,
		Temperature: req.Temperature,
		Humidity:    req.Humidity,
	})
}

// historyHandler 获取历史天气记录
func historyHandler(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if val := r.URL.Query().Get("limit"); val != "" {
		fmt.Sscanf(val, "%d", &limit)
	}

	records, err := weatherDB.GetWeatherHistory(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HistoryResponse{Records: records})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/weather", weatherHandler)
	http.HandleFunc("/history", historyHandler)

	addr := ":" + port
	fmt.Printf("🚀 Server starting on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health      - Health check")
	fmt.Println("  POST /weather     - Save weather data")
	fmt.Println("  GET  /history     - Get historical records")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
