package ml_test

import (
	"testing"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestNewHumidityReference(t *testing.T) {
	hr := ml.NewHumidityReference(1.472, 1.3)
	require.InDelta(t, 1.3, hr.V100(), 1e-9)
	require.InDelta(t, 1.472, hr.V40(), 1e-9)
	require.InDelta(t, 1.3286666666666667, hr.TargetMoistureVoltage(), 1e-9)
}

func TestHumidityReference_LowAndSuspicious(t *testing.T) {
	hr := ml.NewHumidityReference(1.8, 1.2)
	tests := []struct {
		name       string
		voltage    float64
		low        bool
		suspicious bool
	}{
		{name: "50 percent", voltage: 1.7},
		{name: "40 percent", voltage: 1.8, low: true},
		{name: "0 percent", voltage: 2.2, low: true},
		{name: "-30 percent", voltage: 2.5, low: true},
		{name: "-40 percent", voltage: 2.6, low: true, suspicious: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.low, hr.IsLow(tt.voltage))
			require.Equal(t, tt.suspicious, hr.IsSuspicious(tt.voltage))
		})
	}
}
