// Модуль обеспечивает корректную работу с метриками
// Метрики бывают двух типов:
// - Тип gauge (float64) — метрика, новое значение которой полностью замещает текущее значение на сервере.
// - Тип counter (int64) — метрика-счетчик. Агент отправляет дельту, на которую должно измениться значение счетчика за сервере.
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

// Metric описывает метрику, где:
// - ID - уникальный идентификатор метрики.
// - MType - тип метрики (Counter или Gauge).
// - Delta - значение, на которое изменяется метрика типа counter, если метрика типа gauge - не заполнено.
// - Value - значение метрики типа gauge, если метрика типа counter - не заполнено.
// - ValueStr - строковое представление значения метрики.
type Metric struct {
	ID       string   `json:"id"`
	MType    string   `json:"type"`
	Delta    *int64   `json:"delta,omitempty"`
	Value    *float64 `json:"value,omitempty"`
	ValueStr string   `json:"-"`
}

// NewMetrics создает новый экземпляр метрики по идентификатору и типу метрики
func NewMetrics(id string, mType string) *Metric {
	metric := Metric{}
	metric.ID = id
	metric.MType = mType

	return &metric
}

// UpdateValue обновляет значение метрики
func (metric *Metric) UpdateValue(value any) error {
	switch metric.MType {
	case Gauge:
		metricValue, err := GaugeValue(value)
		if err != nil {
			return fmt.Errorf("error converting gauge value: %w", err)
		}
		if metric.Value == nil {
			metric.Value = &metricValue
		} else {
			*metric.Value = metricValue
		}
	case Counter:
		metricsValue, err := CounterValue(value)
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
	metric.ValueStr = metric.getValueString()

	return nil
}

// GetValue возвращает значение метрики. Тип возвращаемого значения зависит от типа метрики
func (metric Metric) GetValue() any {
	switch metric.MType {
	case Gauge:
		return *metric.Value
	case Counter:
		return *metric.Delta
	}
	return nil
}

func (metric Metric) getValueString() string {
	switch metric.MType {
	case Gauge:
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case Counter:
		return strconv.FormatInt(*metric.Delta, 10)
	}
	return ""
}

// Check проверяет корректность заполненности метрики:
// - корректный идентификатор (начинается с буквы, не содержит служебных символов)
// - корректный тип (gauge или counter)
// - заполнено корректное поле значения в зависимости от типа
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

// CompareTypes сравнивает входящий тип с текущим. Если типы не равны - возвращает ошибку.
func (metric Metric) CompareTypes(mType string) error {
	if metric.MType != mType {
		return fmt.Errorf("%w: metric type does not match incoming metric type. expected: %s, have: %s", ErrMetricValidation, metric.MType, mType)
	}
	return nil
}

// GaugeValue преобразует входящий параметр типа any в тип значения gauge (float64).
func GaugeValue(value any) (float64, error) {
	metricsValue, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("value conversion error to float64: %v", value)
	}
	return metricsValue, nil
}

// CounterValue преобразует входящий параметр типа any в тип значения counter (int64).
func CounterValue(value any) (int64, error) {
	metricsValue, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("value conversion error to int64: %v", value)
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

// GetMetricIDs возвращает слайс идентификаторов метрик по входному параметру со слайсом метрик
func GetMetricIDs(metrics []*Metric) []string {
	mNames := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		mNames = append(mNames, metric.ID)
	}
	return mNames
}
