package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_checkMetricID(t *testing.T) {

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{name: "Успешный тест 1", id: "Alloc", want: true},
		{name: "Успешный тест 2", id: "All1oc1", want: true},
		{name: "Некорректный id - начинается с цифры", id: "1Alloc", want: false},
		{name: "Некорректный id - спецсимволы", id: "All*oc", want: false},
		{name: "Некорректный id - пробел в начале", id: " Alloc", want: false},
		{name: "Некорректный id - пробел в конце", id: "Alloc ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkMetricID(tt.id)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
