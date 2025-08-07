package service

import models "github.com/galogen13/yandex-go-metrics/internal/model"

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
	}
	return metrics, nil
}

func NewMetrics(ID string, MType string) models.Metrics {

	metrics := models.Metrics{}
	metrics.ID = ID
	metrics.MType = MType

	return metrics

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
