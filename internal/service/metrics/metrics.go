package metrics

import (
	"fmt"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func NewMetrics(ID string, MType string) Metric {

	metric := Metric{}
	metric.ID = ID
	metric.MType = MType

	return metric

}

func (metric Metric) CheckType(MType string) error {
	if metric.MType != MType {
		return fmt.Errorf("metric type does not match incoming metric type. expected: %s, have: %s", metric.MType, MType)
	}
	return nil
}

func (metric *Metric) UpdateValue(Value any) error {
	switch metric.MType {
	case Gauge:
		metricValue, err := gaugeValue(Value)
		if err != nil {
			return fmt.Errorf("error converting gauge value: %w", err)
		}
		if metric.Value == nil {
			metric.Value = &metricValue
		} else {
			*metric.Value = metricValue
		}
	case Counter:
		metricsValue, err := counterValue(Value)
		if err != nil {
			return fmt.Errorf("error converting counter value: %w", err)
		}
		if metric.Delta == nil {
			metric.Delta = &metricsValue
		} else {
			*metric.Delta += metricsValue
		}
	default:
		return fmt.Errorf("invalid metric type when updating value: %s", metric.MType)
	}

	return nil
}

func (metric Metric) GetValue() any {
	switch metric.MType {
	case Gauge:
		return *metric.Value
	case Counter:
		return *metric.Delta
	}
	return nil
}

func GetMetricsValues(metricsList []Metric) map[string]any {

	result := make(map[string]any)

	for _, metric := range metricsList {
		result[metric.ID] = metric.GetValue()
	}

	return result
}

func gaugeValue(Value any) (float64, error) {
	metricsValue, ok := Value.(float64)
	if !ok {
		return 0, fmt.Errorf("value conversion error to float64: %v", Value)
	}
	return metricsValue, nil
}

func counterValue(Value any) (int64, error) {
	metricsValue, ok := Value.(int64)
	if !ok {
		return 0, fmt.Errorf("value conversion error to int64: %v", Value)
	}
	return metricsValue, nil
}
