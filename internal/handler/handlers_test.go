package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantStruct struct {
	status      int
	response    string
	contentType string
}

func TestUpdateHandler(t *testing.T) {

	stor := storage.NewMemStorage()

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
		{name: "Некорректный context type",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/counter/Counter/1",
			contentType: "test",
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
		{name: "Успешное добавление counter в непустое хранилище",
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
		{name: "Некорректное значение counter",
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
		{name: "Некорректный метод 1",
			storage:     stor,
			method:      http.MethodGet,
			url:         "/update/gauge/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный метод 2",
			storage:     stor,
			method:      http.MethodPut,
			url:         "/update/gauge/Alloc/400",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusMethodNotAllowed, response: "", contentType: respContentTypeTextPlain}},
		{name: "Некорректный тип метрики",
			storage:     stor,
			method:      http.MethodPost,
			url:         "/update/testerr/Counter/1",
			contentType: reqContentTypeTextPlain,
			want:        wantStruct{status: http.StatusBadRequest, response: "", contentType: respContentTypeTextPlain}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.url, nil)
			request.Header.Set("Content-type", test.contentType)
			request.Pattern = "/update/"
			w := httptest.NewRecorder()
			f := UpdateHandler(test.storage)
			f(w, request)

			res := w.Result()
			assert.Equal(t, test.want.status, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.Empty(t, resBody)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_checkMetricsID(t *testing.T) {

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{name: "Успешный тест 1", id: "Alloc", want: true},
		{name: "Успешный тест 2", id: "All1oc1", want: true},
		{name: "Некорректный id - начинается с цифры", id: "1Alloc", want: false},
		{name: "Некорректный id - спецсимволы", id: "All*oc", want: false},
		{name: "Некорректный id - пробел в начале", id: " Alloc", want: false},
		{name: "Некорректный id - пробел в конце", id: "Alloc ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkMetricsID(tt.id); got != tt.want {
				t.Errorf("checkMetricsID() = %v, want %v", got, tt.want)
			}
		})
	}
}
