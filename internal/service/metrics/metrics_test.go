package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricsValue(t *testing.T) {
	type args struct {
		metric Metric
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
		want    Metric
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
