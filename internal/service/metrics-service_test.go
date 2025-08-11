package service

import (
	"testing"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricsValue(t *testing.T) {
	type args struct {
		metrics models.Metrics
		Value   any
	}

	var gaugeValue float64 = 12.34
	newGauge := NewMetrics("Alloc", models.Gauge)
	newGauge.Value = &gaugeValue

	var counterValue int64 = 2
	newCounter := NewMetrics("Counter", models.Counter)
	newCounter.Delta = &counterValue

	tests := []struct {
		name    string
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name: "",
			args: args{
				metrics: NewMetrics("Alloc", models.Gauge),
				Value:   gaugeValue,
			},
			want: newGauge,
		},
		{
			name: "",
			args: args{
				metrics: NewMetrics("Counter", models.Counter),
				Value:   counterValue,
			},
			want: newCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateMetricsValue(tt.args.metrics, tt.args.Value)
			if tt.args.metrics.MType == models.Gauge {
				assert.Equal(t, *tt.want.Value, *got.Value)
				assert.Nil(t, got.Delta)
			}

			if tt.args.metrics.MType == models.Counter {
				assert.Equal(t, *tt.want.Delta, *got.Delta)
				assert.Nil(t, got.Value)
			}

			require.NoError(t, err)
		})
	}
}
