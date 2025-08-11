package service

import models "github.com/galogen13/yandex-go-metrics/internal/model"

func NewMetrics(ID string, MType string) models.Metrics {

	metrics := models.Metrics{}
	metrics.ID = ID
	metrics.MType = MType

	return metrics

}

func CheckMetricsType(metrics models.Metrics, MType string) error {
	if metrics.MType != MType {
		return models.ErrorMetricsIncorrectUse
	}
	return nil
}

func UpdateMetricsValue(metrics models.Metrics, Value any) (models.Metrics, error) {
	switch metrics.MType {
	case models.Gauge:
		metricsValue, err := gaugeValue(Value)
		if err != nil {
			return models.Metrics{}, err
		}
		if metrics.Value == nil {
			metrics.Value = &metricsValue
		} else {
			*metrics.Value = metricsValue
		}
	case models.Counter:
		metricsValue, err := counterValue(Value)
		if err != nil {
			return models.Metrics{}, err
		}
		if metrics.Delta == nil {
			metrics.Delta = &metricsValue
		} else {
			*metrics.Delta += metricsValue
		}
	default:
		return models.Metrics{}, models.ErrorMetricsNotExists
	}
	return metrics, nil
}

func GetMetricsValue(metrics models.Metrics) any {
	switch metrics.MType {
	case models.Gauge:
		return *metrics.Value
	case models.Counter:
		return *metrics.Delta
	}
	return nil
}

func GetMetricsValues(metricsList []models.Metrics) map[string]any {

	result := make(map[string]any)

	for _, metrics := range metricsList {
		result[metrics.ID] = GetMetricsValue(metrics)
	}

	return result
}

func gaugeValue(Value any) (float64, error) {
	metricsValue, ok := Value.(float64)
	if !ok {
		return 0, models.ErrorMetricsIncorrectValue
	}
	return metricsValue, nil
}

func counterValue(Value any) (int64, error) {
	metricsValue, ok := Value.(int64)
	if !ok {
		return 0, models.ErrorMetricsIncorrectValue
	}
	return metricsValue, nil
}
