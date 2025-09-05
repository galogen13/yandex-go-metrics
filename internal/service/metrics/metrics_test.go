package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateValue(t *testing.T) {
	type args struct {
		metric *Metric
		Value  any
	}

	gaugeValue := 12.34
	newGauge := NewMetrics("Alloc", Gauge)
	newGauge.Value = &gaugeValue

	var counterValue int64 = 2
	newCounter := NewMetrics("Counter", Counter)
	newCounter.Delta = &counterValue

	tests := []struct {
		name    string
		args    args
		want    *Metric
		wantErr bool
	}{
		{
			name: "",
			args: args{
				metric: NewMetrics("Alloc", Gauge),
				Value:  gaugeValue,
			},
			want: newGauge,
		},
		{
			name: "",
			args: args{
				metric: NewMetrics("Counter", Counter),
				Value:  counterValue,
			},
			want: newCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.metric.UpdateValue(tt.args.Value)
			if tt.args.metric.MType == Gauge {
				assert.Equal(t, *tt.want.Value, *tt.args.metric.Value)
				assert.Nil(t, tt.args.metric.Delta)
			}

			if tt.args.metric.MType == Counter {
				assert.Equal(t, *tt.want.Delta, *tt.args.metric.Delta)
				assert.Nil(t, tt.args.metric.Value)
			}

			require.NoError(t, err)
		})
	}
}

func Test_Check(t *testing.T) {

	value := 0.31
	delta := int64(14)

	tests := []struct {
		name    string
		metric  Metric
		wantErr bool
	}{
		{name: "Успешный тест 1",
			metric:  Metric{ID: "Alloc", MType: Gauge, Value: &value},
			wantErr: false},
		{name: "Успешный тест 2",
			metric:  Metric{ID: "Counter", MType: Counter, Delta: &delta},
			wantErr: false},
		{name: "Некорректный id - начинается с цифры",
			metric:  Metric{ID: "1Alloc", MType: Gauge, Value: &value},
			wantErr: true},
		{name: "Некорректный id - спецсимволы",
			metric:  Metric{ID: "Alloc*", MType: Gauge, Value: &value},
			wantErr: true},
		{name: "Некорректный id - пробел в начале",
			metric:  Metric{ID: " Alloc", MType: Gauge, Value: &value},
			wantErr: true},
		{name: "Некорректный id - пробел в конце",
			metric:  Metric{ID: "Alloc ", MType: Gauge, Value: &value},
			wantErr: true},
		{name: "Некорректный тип",
			metric:  Metric{ID: "Alloc", MType: "test", Value: &value},
			wantErr: true},
		{name: "Некорректное поле значения для Gauge",
			metric:  Metric{ID: "Alloc", MType: Gauge, Delta: &delta},
			wantErr: true},
		{name: "Некорректное поле значения для Counter",
			metric:  Metric{ID: "Counter", MType: Counter, Value: &value},
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Check(true)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrMetricValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
