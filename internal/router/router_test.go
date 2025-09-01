package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	storage "github.com/galogen13/yandex-go-metrics/internal/repository"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	respContentTypeTextHTML = "text/html; charset=utf-8"
)

type testRequestResponse struct {
	StatusCode        int
	Body, ContentType string
}

func TestRouter_Update(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}

	serverService := server.NewServerService(config, stor)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     server.Storage
		method      string
		url         string
		body        string
		contentType string
		want        wantStruct
	}{
		{name: "Успешное добавление gauge в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":200}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":400}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Повторное добавление в непустое хранилище с некорректным типом",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"counter","value":600}`,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter","delta":1}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный url - нет типа метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","value":200}`,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректное значение",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter","delta":sdf}`,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректное значение для counter",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter","delta":0.01}`,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный url",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/updateeeeee",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":200}`,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный тип метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counterrra","delta":1}`,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный метод: GET вместо POST",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":400}`,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный метод 2: PUT вместо POST",
			storage:     stor,
			method:      http.MethodPut,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":400}`,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, test.method, test.url, test.contentType, test.body)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_Get(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}

	serverService := server.NewServerService(config, stor)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     server.Storage
		method      string
		url         string
		contentType string
		body        string
		want        wantStruct
	}{
		{name: "Успешное добавление gauge в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge","value":20.99}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения gauge",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"gauge"}`,
			want:        wantStruct{status: http.StatusOK, response: `{"id":"Alloc","type":"gauge","value":20.99}`, contentType: "application/json"}},
		{name: "Некорректный тип метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"counter"}`,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Несуществующая метрика",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			body:        `{"id":"Malloc","type":"gauge"}`,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter","delta":12}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения counter",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter"}`,
			want:        wantStruct{status: http.StatusOK, response: `{"id":"Counter","type":"counter","delta":12}`, contentType: "application/json"}},
		{name: "Успешное добавление counter в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter","delta":5}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения counter",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			body:        `{"id":"Counter","type":"counter"}`,
			want:        wantStruct{status: http.StatusOK, response: `{"id":"Counter","type":"counter","delta":17}`, contentType: "application/json"}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, test.method, test.url, test.contentType, test.body)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_UpdateURL(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}

	serverService := server.NewServerService(config, stor)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     server.Storage
		method      string
		url         string
		contentType string
		want        wantStruct
	}{
		{name: "Успешное добавление gauge в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/gauge/Alloc/200",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/gauge/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Повторное добавление в непустое хранилище с некорректным типом",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/1",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный url - нет типа метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/Counter/1",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректное значение",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/sdf",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректное значение для counter",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/0.01",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный url",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/updateeeeee/counter/Counter/1",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный тип метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counterrra/Counter/1",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный метод: GET вместо POST",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/update/gauge/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный метод 2: PUT вместо POST",
			storage:     stor,
			method:      http.MethodPut,
			url:         "/update/gauge/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, test.method, test.url, test.contentType, "")
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_GetList(t *testing.T) {

	type wantStruct struct {
		status         int
		tdMetricsID    string
		tdMetricsValue string
		contentType    string
	}

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}

	serverService := server.NewServerService(config, stor)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     server.Storage
		method      string
		url         string
		contentType string
		want        wantStruct
	}{
		{name: "Получение страницы с пустым хранилищем",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, tdMetricsID: "", tdMetricsValue: "", contentType: respContentTypeTextHTML}},
		{name: "Добавление gauge для теста",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/gauge/Alloc/100.2",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, tdMetricsID: "", tdMetricsValue: "", contentType: respContentTypeTextPlain}},
		{name: "Проверка counter для теста ",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/2",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, tdMetricsID: "", tdMetricsValue: "", contentType: respContentTypeTextPlain}},
		{name: "Получение страницы с непустым хранилищем",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, tdMetricsID: "<td>Counter</td>", tdMetricsValue: "<td>2</td>", contentType: respContentTypeTextHTML}},
		{name: "Получение страницы с непустым хранилищем",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, tdMetricsID: "<td>Alloc</td>", tdMetricsValue: "<td>100.2</td>", contentType: respContentTypeTextHTML}},
	}

	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, test.method, test.url, test.contentType, "")
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)

		assert.Contains(t, resp.Body, test.want.tdMetricsID)
		assert.Contains(t, resp.Body, test.want.tdMetricsValue)

		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_GetURL(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}

	serverService := server.NewServerService(config, stor)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     server.Storage
		method      string
		url         string
		contentType string
		want        wantStruct
	}{
		{name: "Успешное добавление gauge в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/gauge/Alloc/20.99",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения gauge",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/gauge/Alloc",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "20.99", contentType: respContentTypeTextPlain}},
		{name: "Некорректный тип метрики",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/counter/Alloc",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Несуществующая метрика",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/gauge/Malloc",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/12",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения counter",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/counter/Counter",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "12", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/18",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения counter",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/counter/Counter",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusOK, response: "30", contentType: respContentTypeTextPlain}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, test.method, test.url, test.contentType, "")
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, contentType string, body string) testRequestResponse {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	result := testRequestResponse{
		StatusCode:  resp.StatusCode,
		Body:        string(respBodyBytes),
		ContentType: resp.Header.Get("Content-Type"),
	}

	return result
}
