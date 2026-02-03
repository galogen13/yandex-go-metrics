package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	addinfo "github.com/galogen13/yandex-go-metrics/internal/service/additional-info"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

// mockServer реализует интерфейс handler.Server для тестирования.
type mockServer struct{}

func (m *mockServer) UpdateMetric(ctx context.Context, metric *metrics.Metric, addInfo addinfo.AddInfo) error {
	return nil
}

func (m *mockServer) UpdateMetrics(ctx context.Context, metrics []*metrics.Metric, addInfo addinfo.AddInfo) error {
	return nil
}

func (m *mockServer) GetMetric(ctx context.Context, metric *metrics.Metric) (*metrics.Metric, error) {
	if metric.ID == "Alloc" && metric.MType == metrics.Gauge {
		metric := metrics.NewMetrics("Alloc", metrics.Gauge)
		metric.UpdateValue(123.45)
		return metric, nil
	}
	return nil, metrics.ErrMetricNotFound
}

func (m *mockServer) GetAllMetrics(ctx context.Context) ([]*metrics.Metric, error) {
	return []*metrics.Metric{
		{ID: "Alloc", MType: metrics.Gauge, Value: func() *float64 { v := 123.45; return &v }()},
		{ID: "PollCount", MType: metrics.Counter, Delta: func() *int64 { v := int64(42); return &v }()},
	}, nil
}

func (m *mockServer) PingStorage(ctx context.Context) error {
	return nil
}

func (m *mockServer) Key() string {
	return "test-key"
}

// Example_metricsRouter демонстрирует создание и использование роутера метрик.
func Example_metricsRouter() {
	// Создаем mock сервер
	server := &mockServer{}

	// Создаем роутер
	router := metricsRouter(server)

	// Создаем тестовый сервер с роутером
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Тестируем endpoint для получения списка метрик
	resp, err := http.Get(testServer.URL + "/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Проверяем, что ответ содержит HTML с метриками
	hasHTML := strings.Contains(string(body), "<html>") ||
		strings.Contains(string(body), "<table>")

	fmt.Printf("Status: %d, Is HTML: %v\n", resp.StatusCode, hasHTML)
	// Output:
	// Status: 200, Is HTML: true
}

// Example_metricsRouter_ping демонстрирует использование endpoint /ping.
func Example_metricsRouter_ping() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Тестируем ping endpoint
	resp, err := http.Get(testServer.URL + "/ping")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Ping Status: %d\n", resp.StatusCode)
	// Output:
	// Ping Status: 200
}

// Example_metricsRouter_update демонстрирует обновление метрики через URL.
func Example_metricsRouter_update() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Тестируем обновление метрики через URL (старый формат)
	resp, err := http.Post(testServer.URL+"/update/gauge/Alloc/123.45", "", nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Update URL Status: %d\n", resp.StatusCode)
	// Output:
	// Update URL Status: 200
}

// Example_metricsRouter_update_json демонстрирует обновление метрики в формате JSON.
func Example_metricsRouter_update_json() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Подготавливаем JSON запрос
	jsonStr := `{"id":"Alloc","type":"gauge","value":123.45}`

	resp, err := http.Post(
		testServer.URL+"/update",
		"application/json",
		bytes.NewBufferString(jsonStr),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Update JSON Status: %d\n", resp.StatusCode)
	// Output:
	// Update JSON Status: 200
}

// Example_metricsRouter_value демонстрирует получение значения метрики через URL.
func Example_metricsRouter_value() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Тестируем получение значения метрики через URL
	resp, err := http.Get(testServer.URL + "/value/gauge/Alloc")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Value URL Status: %d, Value: %s\n",
		resp.StatusCode, strings.TrimSpace(string(body)))
	// Output:
	// Value URL Status: 200, Value: 123.45
}

// Example_metricsRouter_value_json демонстрирует получение значения метрики в формате JSON.
func Example_metricsRouter_value_json() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Подготавливаем JSON запрос
	jsonStr := `{"id":"Alloc","type":"gauge"}`

	resp, err := http.Post(
		testServer.URL+"/value",
		"application/json",
		bytes.NewBufferString(jsonStr),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Value JSON Status: %d, Response: %s\n",
		resp.StatusCode, string(body))
	// Output:
	// Value JSON Status: 200, Response: {"id":"Alloc","type":"gauge","value":123.45}
}

// Example_metricsRouter_updates демонстрирует массовое обновление метрик.
func Example_metricsRouter_updates() {
	server := &mockServer{}
	router := metricsRouter(server)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Подготавливаем JSON запрос с массивом метрик
	jsonStr := `[
		{"id":"Alloc","type":"gauge","value":123.45},
		{"id":"PollCount","type":"counter","delta":1}
	]`

	resp, err := http.Post(
		testServer.URL+"/updates",
		"application/json",
		bytes.NewBufferString(jsonStr),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Updates Status: %d\n", resp.StatusCode)
	// Output:
	// Updates Status: 200
}
