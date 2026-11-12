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
		{name: "with a forecast at the threshold, then it returns true", forecast: new(25.0), sensor: new(20.0), expected: true},
		{name: "with a forecast below the threshold and a hot sensor, then it returns false", forecast: new(24.9), sensor: new(30.0)},
		{name: "without forecast and a sensor at the threshold, then it returns true", sensor: new(31.0), expected: true},
		{name: "without forecast and a sensor below the threshold, then it returns false", sensor: new(30.9)},
		{name: "without any temperature, then it returns true", expected: true},
	}
	for _, tt := range tests {
		t.Run(`Given an AmbientTemperature struct,
		when IsHot method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, ml.NewAmbientTemperature(tt.forecast, tt.sensor).IsHot())
		})
	}
}
