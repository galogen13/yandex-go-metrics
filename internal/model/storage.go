package models

import "github.com/galogen13/yandex-go-metrics/internal/service/metrics"

type Storage interface {
	Update(metric metrics.Metric)
	Get(ID string) (bool, metrics.Metric)
	GetAll() []metrics.Metric
}
