package metrics

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrMetricValidation = errors.New("metric validation error")
	ErrMetricNotFound   = errors.New("metric not found")
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
	Hash  string   `json:"-"` //`json:"hash,omitempty"`
}

func NewMetrics(ID string, MType string) *Metric {
	metric := Metric{}
	metric.ID = ID
	metric.MType = MType

	return &metric
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

func (metric Metric) GetValueString() string {
	switch metric.MType {
	case Gauge:
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case Counter:
		return strconv.FormatInt(*metric.Delta, 10)
	}
	return ""
}

func GetMetricsValues(metricsList []*Metric) map[string]any {

	result := make(map[string]any)

	for _, metric := range metricsList {
		result[metric.ID] = metric.GetValue()
	}

	return result
}

func GetMetricsFromMap(metrics map[string]*Metric) []*Metric {

	result := make([]*Metric, 0, len(metrics))

	for _, metric := range metrics {
		result = append(result, metric)
	}

	return result
}

func (metric Metric) Check(checkValue bool) error {
	metricIDIsCorrect := metric.checkID()
	if !metricIDIsCorrect {
		return fmt.Errorf("%w: metric ID is incorrect: %s", ErrMetricValidation, metric.ID)
	}

	if !metric.checkType() {
		return fmt.Errorf("%w: metric type is incorrect: %s", ErrMetricValidation, metric.MType)
	}

	if checkValue {
		if !metric.checkValue() {
			return fmt.Errorf("%w: metric value is incorrect: MType: %s, Delta: %v, Value: %v", ErrMetricValidation, metric.MType, metric.Delta, metric.Value)
		}
	}
	return nil
}

func (metric Metric) CompareTypes(MType string) error {
	if metric.MType != MType {
		return fmt.Errorf("%w: metric type does not match incoming metric type. expected: %s, have: %s", ErrMetricValidation, metric.MType, MType)
	}
	return nil
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

func (metric Metric) checkType() bool {
	return metric.MType == Counter || metric.MType == Gauge
}

var metricIDRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*$")

func (metric Metric) checkID() bool {
	match := metricIDRegex.MatchString(metric.ID)
	return match
}

func (metric Metric) checkValue() bool {
	switch metric.MType {
	case Gauge:
		return metric.Value != nil
	case Counter:
		return metric.Delta != nil
	}
	return false
}
