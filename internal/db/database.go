package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// WeatherRecord 天气记录模型
type WeatherRecord struct {
	ID          int64     `json:"id"`
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	RecordedAt  time.Time `json:"recorded_at"`
}

// Database 数据库连接封装
type Database struct {
	db *sql.DB
}

// New 创建新的数据库实例
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}

	// 初始化表结构
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// initSchema 创建 weather_records 表
func (d *Database) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS weather_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		city TEXT NOT NULL,
		temperature REAL NOT NULL,
		humidity INTEGER NOT NULL,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := d.db.Exec(query)
	return err
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveWeather 保存天气记录
func (d *Database) SaveWeather(city string, temperature float64, humidity int) (int64, error) {
	if city == "" {
		return 0, fmt.Errorf("city name cannot be empty")
	}
	query := `INSERT INTO weather_records (city, temperature, humidity) VALUES (?, ?, ?)`
	result, err := d.db.Exec(query, city, temperature, humidity)
	if err != nil {
		return 0, fmt.Errorf("failed to save weather record: %w", err)
	}

	id, _ := result.LastInsertId()
	return id, nil
}

// GetWeatherHistory 获取历史天气记录
func (d *Database) GetWeatherHistory(limit int) ([]WeatherRecord, error) {
	query := `SELECT id, city, temperature, humidity, recorded_at FROM weather_records ORDER BY recorded_at DESC LIMIT ?`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query weather history: %w", err)
	}
	defer rows.Close()

	var records []WeatherRecord
	for rows.Next() {
		var r WeatherRecord
		if err := rows.Scan(&r.ID, &r.City, &r.Temperature, &r.Humidity, &r.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, r)
	}

	return records, nil
}
