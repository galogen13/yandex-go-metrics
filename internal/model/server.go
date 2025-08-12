package models

type Server interface {
	UpdateMetric(ID string, MType string, Value any) error
	GetMetricValue(ID string, MType string) (any, error)
	GetAllMetricsValues() map[string]any
	Host() string
}
