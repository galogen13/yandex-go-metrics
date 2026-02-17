package pool_test

import (
	"testing"

	"github.com/galogen13/yandex-go-metrics/internal/pool"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/stretchr/testify/assert"
)

func newMetricFunc() *metrics.Metric {
	return &metrics.Metric{}
}

func TestNewPool(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		p, err := pool.New(newMetricFunc)
		assert.Nil(t, err)
		assert.NotNil(t, p)
	})

	t.Run("nil newFunc", func(t *testing.T) {
		p, err := pool.New[*metrics.Metric](nil)
		assert.NotNil(t, err)
		assert.Nil(t, p)
	})
}

func TestPoolGet(t *testing.T) {
	t.Run("get returns non-nil object", func(t *testing.T) {
		p, err := pool.New(newMetricFunc)
		assert.Nil(t, err)

		obj := p.Get()
		assert.NotNil(t, obj)
	})

}

func TestPoolPut(t *testing.T) {
	t.Run("put returns object and reset is called", func(t *testing.T) {

		newMetric := metrics.NewMetrics("Alloc", metrics.Gauge)
		newMetric.UpdateValue(1.2)

		p, err := pool.New(newMetricFunc)

		assert.Nil(t, err)

		p.Put(newMetric)

		resettedMetric := p.Get()

		assert.Equal(t, resettedMetric.ID, "")
		assert.Equal(t, resettedMetric.MType, "")
		assert.Equal(t, *resettedMetric.Value, 0.0)
		assert.Nil(t, resettedMetric.Delta)
		assert.Equal(t, resettedMetric.ValueStr, "")

	})
}

func TestPoolReuse(t *testing.T) {
	t.Run("objects are reused after put", func(t *testing.T) {
		p, err := pool.New(newMetricFunc)
		assert.Nil(t, err)

		obj1 := p.Get()
		obj1.ID = "Alloc"
		obj1.MType = metrics.Gauge
		obj1.UpdateValue(1.2)
		p.Put(obj1)

		obj2 := p.Get()

		assert.Equal(t, obj1, obj2)

		assert.Equal(t, obj2.ID, "")
		assert.Equal(t, obj2.MType, "")
		assert.Equal(t, *obj2.Value, 0.0)
		assert.Nil(t, obj2.Delta)
		assert.Equal(t, obj2.ValueStr, "")
	})

	t.Run("objects are not reused if not put back", func(t *testing.T) {
		p, err := pool.New(newMetricFunc)
		assert.Nil(t, err)

		obj1 := p.Get()
		obj1.ID = "Alloc"

		obj2 := p.Get()

		assert.NotEqual(t, obj1, obj2)
	})
}
