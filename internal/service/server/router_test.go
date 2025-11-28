package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/galogen13/yandex-go-metrics/internal/audit"
	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	storage "github.com/galogen13/yandex-go-metrics/internal/repository/memstorage"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
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

type testCase struct {
	name        string
	storage     Storage
	method      string
	url         string
	contentType string
	body        string
	compressReq bool
	want        wantStruct
}

type wantStruct struct {
	status         int
	response       string
	contentType    string
	tdMetricsID    string
	tdMetricsValue string
}

func TestRouter_Update(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
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
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_Get(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
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
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_Compression(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
		{name: "Успешное добавление gauge в пустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			compressReq: true,
			body:        `{"id":"Alloc","type":"gauge","value":20.99}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения gauge",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			compressReq: true,
			body:        `{"id":"Alloc","type":"gauge"}`,
			want:        wantStruct{status: http.StatusOK, response: `{"id":"Alloc","type":"gauge","value":20.99}`, contentType: "application/json"}},
		{name: "Успешное добавление counter в непустое хранилище",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update",
			contentType: "application/json",
			compressReq: true,
			body:        `{"id":"Counter","type":"counter","delta":5}`,
			want:        wantStruct{status: http.StatusOK, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное получение значения counter",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/value",
			contentType: "application/json",
			compressReq: true,
			body:        `{"id":"Counter","type":"counter"}`,
			want:        wantStruct{status: http.StatusOK, response: `{"id":"Counter","type":"counter","delta":5}`, contentType: "application/json"}},
	}
	for _, test := range tests {
		// Тесты выполняются последовательно, не в отдельных горутинах, т.к. результат прошлых кейсов влияет на будущие
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_UpdateURL(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
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
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_GetList(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
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
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)

		assert.Contains(t, resp.Body, test.want.tdMetricsID)
		assert.Contains(t, resp.Body, test.want.tdMetricsValue)

		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func TestRouter_GetURL(t *testing.T) {

	stor := storage.NewMemStorage()
	config := config.ServerConfig{Host: "localhost:8080"}
	auditService := audit.NewAuditServise()

	serverService := NewServerService(&config, stor, auditService)

	ts := httptest.NewServer(metricsRouter(serverService))
	defer ts.Close()

	tests := []testCase{
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
		resp := testRequest(t, ts, &test)
		assert.Equal(t, test.want.status, resp.StatusCode, test.name)
		assert.Equal(t, test.want.response, resp.Body, test.name)
		assert.Equal(t, test.want.contentType, resp.ContentType, test.name)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, tc *testCase) testRequestResponse {

	var body io.Reader
	var err error

	if tc.compressReq {
		compressedBody, err := compressBody(tc.body)
		require.NoError(t, err)
		body = compressedBody
	} else {
		body = strings.NewReader(tc.body)
	}
	req, err := http.NewRequestWithContext(t.Context(), tc.method, ts.URL+tc.url, body)
	req.Header.Set("Content-Type", tc.contentType)
	if tc.compressReq {
		req.Header.Set("Content-Encoding", "gzip")
	}
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

func compressBody(data string) (io.Reader, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write([]byte(data)); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func TestGzipCompression(t *testing.T) {

	id := "Alloc"
	mType := metrics.Gauge
	value := 20.99
	stor := storage.NewMemStorage()
	metric := metrics.NewMetrics(id, mType)
	err := metric.UpdateValue(value)
	require.NoError(t, err)
	err = stor.Update(context.Background(), []*metrics.Metric{metric})
	require.NoError(t, err)
	config, err := config.GetServerConfig()
	require.NoError(t, err)
	auditService := audit.NewAuditServise()

	serverService := NewServerService(config, stor, auditService)

	handler := http.HandlerFunc(compression.GzipMiddleware(handler.GetValueHandler(serverService)))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := fmt.Sprintf(`{"id":%q,"type":%q}`, id, mType)

	successBody := fmt.Sprintf(`{"id":%q,"type":%q,"value":%.2f}`, id, mType, value)

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
