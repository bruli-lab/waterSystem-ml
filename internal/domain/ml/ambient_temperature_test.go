package ml_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

func TestAmbientTemperature_IsHot(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		forecast, sensor *float64
		expected         bool
	}{
		{name: "forecast at threshold is hot", forecast: new(25.0), sensor: new(20.0), expected: true},
		{name: "forecast below threshold is not hot even if the sensor is hot", forecast: new(24.9), sensor: new(30.0)},
		{name: "without forecast the sensor at threshold is hot", sensor: new(31.0), expected: true},
		{name: "without forecast the sensor below threshold is not hot", sensor: new(30.9)},
		{name: "without any temperature it is considered hot", expected: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, ml.NewAmbientTemperature(tt.forecast, tt.sensor).IsHot())
		})
	}
}
