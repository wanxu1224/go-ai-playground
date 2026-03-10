package db

import (
	"os"
	"testing"
	"time"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) (*Database, func()) {
	// 创建临时数据库文件
	tmpfile, err := os.CreateTemp("", "weather_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	dbPath := tmpfile.Name()
	tmpfile.Close()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// 清理函数
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

// ============ 单元测试 ============

// TestNew 测试数据库初始化
func TestNew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}
	if db.db == nil {
		t.Fatal("Expected non-nil sql.DB")
	}
}

// TestSaveWeather 测试保存天气记录
func TestSaveWeather(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		city        string
		temperature float64
		humidity    int
		wantErr     bool
	}{
		{"Beijing", "beijing", 25.5, 60, false},
		{"Shanghai", "shanghai", 28.3, 75, false},
		{"Negative Temp", "beijing", -5.0, 30, false},
		{"Empty City", "", 20.0, 50, true}, // 空城市名应该失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := db.SaveWeather(tt.city, tt.temperature, tt.humidity)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveWeather() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id == 0 {
				t.Error("Expected non-zero ID for successful insert")
			}
		})
	}
}

// TestGetWeatherHistory 测试获取历史记录
func TestGetWeatherHistory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 插入测试数据
	testData := []struct {
		city        string
		temperature float64
		humidity    int
	}{
		{"beijing", 25.5, 60},
		{"shanghai", 28.3, 75},
		{"guangzhou", 30.1, 80},
	}

	for _, td := range testData {
		db.SaveWeather(td.city, td.temperature, td.humidity)
		time.Sleep(10 * time.Millisecond) // 确保时间戳有差异
	}

	// 测试获取全部记录
	records, err := db.GetWeatherHistory(10)
	if err != nil {
		t.Fatalf("GetWeatherHistory() error = %v", err)
	}
	if len(records) != len(testData) {
		t.Errorf("Expected %d records, got %d", len(testData), len(records))
	}

	// 测试限制数量
	records, err = db.GetWeatherHistory(2)
	if err != nil {
		t.Fatalf("GetWeatherHistory(limit=2) error = %v", err)
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records with limit=2, got %d", len(records))
	}

	// 验证记录内容
	for _, r := range records {
		if r.City == "" {
			t.Error("City should not be empty")
		}
		if r.Humidity < 0 || r.Humidity > 100 {
			t.Errorf("Humidity out of range: %d", r.Humidity)
		}
		if r.RecordedAt.IsZero() {
			t.Error("RecordedAt should be set")
		}
	}
}

// TestGetWeatherHistory_Empty 测试空数据库
func TestGetWeatherHistory_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	records, err := db.GetWeatherHistory(10)
	if err != nil {
		t.Fatalf("GetWeatherHistory() error = %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty database, got %d", len(records))
	}
}

// TestDatabase_Close 测试关闭数据库
func TestDatabase_Close(t *testing.T) {
	db, _ := setupTestDB(t)

	err := db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	// 注意：sqlite3 多次 Close 可能不报错，这里只测试第一次关闭成功
}

// TestSaveAndGet_Integration 集成测试：保存后立即查询
func TestSaveAndGet_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 保存记录
	id, err := db.SaveWeather("beijing", 22.5, 55)
	if err != nil {
		t.Fatalf("SaveWeather() error = %v", err)
	}

	// 查询记录
	records, err := db.GetWeatherHistory(1)
	if err != nil {
		t.Fatalf("GetWeatherHistory() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// 验证数据
	r := records[0]
	if r.ID != id {
		t.Errorf("ID mismatch: got %d, want %d", r.ID, id)
	}
	if r.City != "beijing" {
		t.Errorf("City mismatch: got %q, want %q", r.City, "beijing")
	}
	if r.Temperature != 22.5 {
		t.Errorf("Temperature mismatch: got %v, want %v", r.Temperature, 22.5)
	}
	if r.Humidity != 55 {
		t.Errorf("Humidity mismatch: got %d, want %d", r.Humidity, 55)
	}
}

// ============ 基准测试 ============

// BenchmarkSaveWeather 基准测试：保存操作性能
func BenchmarkSaveWeather(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.SaveWeather("beijing", 25.0, 60)
	}
}

// BenchmarkGetWeatherHistory 基准测试：查询操作性能
func BenchmarkGetWeatherHistory(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	// 预填充数据
	for i := 0; i < 100; i++ {
		db.SaveWeather("beijing", 25.0, 60)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetWeatherHistory(10)
	}
}
