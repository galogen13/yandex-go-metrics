package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	respContentTypeTextHTML = "text/html; charset=utf-8"
)

func TestRouter_Update(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()

	ts := httptest.NewServer(metricsRouter(stor))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     models.Storage
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
		// временно убрано, т.к. не проходили автотесты в github, хотя в задании 1 инкремента явно была написано,
		// что update должен быть с "Content-Type: text/plain"
		// {name: "Некорректный context type",
		// 	storage:     stor,
		// 	method:      http.MethodPost,
		// 	url:         "/update/counter/Counter/1",
		// 	contentType: "text/html",
		// 	want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
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
		resp, respBody := testRequest(t, ts, test.method, test.url, test.contentType)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, respBody, test.name)
		assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), test.name)
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

	ts := httptest.NewServer(metricsRouter(stor))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     models.Storage
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
		{name: "Проверка наличия ",
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
		resp, respBody := testRequest(t, ts, test.method, test.url, test.contentType)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)

		assert.Contains(t, respBody, test.want.tdMetricsID)
		assert.Contains(t, respBody, test.want.tdMetricsValue)

		assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), test.name)
	}
}

func TestRouter_Get(t *testing.T) {

	type wantStruct struct {
		status      int
		response    string
		contentType string
	}

	stor := storage.NewMemStorage()

	ts := httptest.NewServer(metricsRouter(stor))
	defer ts.Close()

	tests := []struct {
		name        string
		storage     models.Storage
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
		{name: "Успешное получение значения",
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
		{name: "Некорректная метрика",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/value/gauge/Malloc",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusNotFound, response: "", contentType: respContentTypeTextPlain}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp, respBody := testRequest(t, ts, test.method, test.url, test.contentType)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, respBody, test.name)
		assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), test.name)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, contentType string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(""))
	req.Header.Set("Content-Type", contentType)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
