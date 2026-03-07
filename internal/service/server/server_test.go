package server

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"testing"

	_ "net/http/pprof"

	"github.com/galogen13/yandex-go-metrics/internal/audit"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/repository/memstorage"
	addinfo "github.com/galogen13/yandex-go-metrics/internal/service/additional-info"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	mock_server "github.com/galogen13/yandex-go-metrics/internal/service/server/mocks"
	"go.uber.org/mock/gomock"
)

func BenchmarkUpdateMetrics(b *testing.B) {

	config := &config.ServerConfig{}

	ctrl := gomock.NewController(b)
	mockStorage := mock_server.NewMockStorage(ctrl)
	defer ctrl.Finish()

	as := audit.NewAuditService()

	ctx := context.Background()

	ai := addinfo.AddInfo{}

	ss, _ := NewServerService(config, mockStorage, as)

	metricsCount := 1000

	incomingMetrics := make([]*metrics.Metric, 0, metricsCount)
	existedMetrics := make(map[string]*metrics.Metric)
	metricIDs := make([]string, 0, metricsCount)
	for j := 0; j < metricsCount; j++ {
		var metric *metrics.Metric
		mID := "Test" + strconv.Itoa(j)
		if j%2 == 0 { // каждый второй - gauge
			metric = metrics.NewMetrics(mID, metrics.Gauge)
			err := metric.UpdateValue(rand.Float64())
			if err != nil {
				b.Fatal(err)
			}
		} else {
			metric = metrics.NewMetrics(mID, metrics.Counter)
			err := metric.UpdateValue(rand.Int63n(100))
			if err != nil {
				b.Fatal(err)
			}
		}

		metricIDs = append(metricIDs, mID)

		incomingMetrics = append(incomingMetrics, metric)

		// каждый третий - новая метрика
		if j%3 != 0 {
			existedMetrics[metric.ID] = metric
		}

	}

	mockStorage.EXPECT().GetByIDs(gomock.Any(), metricIDs).Return(existedMetrics, nil).AnyTimes()
	mockStorage.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStorage.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	b.ReportAllocs()

	for b.Loop() {
		ss.UpdateMetrics(ctx, incomingMetrics, ai)
	}

}

// для эксперимента - бенчмарк без использования моков
func BenchmarkUpdateMetrics_MemStorage(b *testing.B) {

	config := &config.ServerConfig{}

	memstorage := memstorage.NewMemStorage()
	as := audit.NewAuditService()

	ctx := context.Background()

	ai := addinfo.AddInfo{}

	ss, _ := NewServerService(config, memstorage, as)

	metricsCount := 1000

	incomingMetrics := make([]*metrics.Metric, 0, metricsCount)
	for j := 0; j < metricsCount; j++ {
		var metric *metrics.Metric
		mID := "Test" + strconv.Itoa(j)
		if j%2 == 0 { // каждый второй - gauge
			metric = metrics.NewMetrics(mID, metrics.Gauge)
			err := metric.UpdateValue(rand.Float64())
			if err != nil {
				b.Fatal(err)
			}
		} else {
			metric = metrics.NewMetrics(mID, metrics.Counter)
			err := metric.UpdateValue(rand.Int63n(100))
			if err != nil {
				b.Fatal(err)
			}
		}

		incomingMetrics = append(incomingMetrics, metric)

		// каждый третий - новая метрика
		if j%3 != 0 {
			ss.UpdateMetric(ctx, metric, ai)
		}

	}

	b.ReportAllocs()

	for b.Loop() {
		ss.UpdateMetrics(ctx, incomingMetrics, ai)
	}

}

func BenchmarkGetMetric(b *testing.B) {

	config := &config.ServerConfig{}

	ctrl := gomock.NewController(b)
	mockStorage := mock_server.NewMockStorage(ctrl)
	defer ctrl.Finish()

	as := audit.NewAuditService()

	ctx := context.Background()

	ss, _ := NewServerService(config, mockStorage, as)

	metricsCount := 100

	incomingMetrics := make([]*metrics.Metric, 0, metricsCount)
	for j := 0; j < metricsCount; j++ {
		var metric *metrics.Metric
		mID := "Test" + strconv.Itoa(j)
		if j%2 == 0 { // каждый второй - gauge
			metric = metrics.NewMetrics(mID, metrics.Gauge)
			err := metric.UpdateValue(rand.Float64())
			if err != nil {
				b.Fatal(err)
			}
		} else {
			metric = metrics.NewMetrics(mID, metrics.Counter)
			err := metric.UpdateValue(rand.Int63n(100))
			if err != nil {
				b.Fatal(err)
			}
		}

		incomingMetrics = append(incomingMetrics, metric)

		// каждая 10 - ошибка
		if j%10 == 0 {
			mockStorage.EXPECT().Get(gomock.Any(), metric).Return(nil, errors.New("test error")).AnyTimes()
		} else if j%3 == 0 { // каждая 3 - не найдена
			mockStorage.EXPECT().Get(gomock.Any(), metric).Return(nil, nil).AnyTimes()
		} else {
			mockStorage.EXPECT().Get(gomock.Any(), metric).Return(metric, nil).AnyTimes()
		}

	}

	b.ReportAllocs()

	for b.Loop() {
		for _, metric := range incomingMetrics {
			ss.GetMetric(ctx, metric)
		}
	}

}

func BenchmarkGetAllMetricsValues(b *testing.B) {

	config := &config.ServerConfig{}

	ctrl := gomock.NewController(b)
	mockStorage := mock_server.NewMockStorage(ctrl)
	defer ctrl.Finish()

	as := audit.NewAuditService()

	ctx := context.Background()

	ss, _ := NewServerService(config, mockStorage, as)

	metricsCount := 1000

	incomingMetrics := make([]*metrics.Metric, 0, metricsCount)
	for j := 0; j < metricsCount; j++ {
		var metric *metrics.Metric
		mID := "Test" + strconv.Itoa(j)
		if j%2 == 0 { // каждый второй - gauge
			metric = metrics.NewMetrics(mID, metrics.Gauge)
			err := metric.UpdateValue(rand.Float64())
			if err != nil {
				b.Fatal(err)
			}
		} else {
			metric = metrics.NewMetrics(mID, metrics.Counter)
			err := metric.UpdateValue(rand.Int63n(100))
			if err != nil {
				b.Fatal(err)
			}
		}

		incomingMetrics = append(incomingMetrics, metric)
	}
	mockStorage.EXPECT().GetAll(gomock.Any()).Return(incomingMetrics, nil).AnyTimes()

	b.ReportAllocs()

	for b.Loop() {
		ss.GetAllMetrics(ctx)
	}

}
