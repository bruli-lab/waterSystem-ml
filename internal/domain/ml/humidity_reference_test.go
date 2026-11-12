package ml_test

import (
	"testing"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestNewHumidityReference(t *testing.T) {
	t.Parallel()

	t.Run(`Given V40 and V100 voltages,
		when the constructor is called, then it returns the voltages and the 90 percent target`, func(t *testing.T) {
		t.Parallel()
		hr := ml.NewHumidityReference(1.472, 1.3)
		require.InDelta(t, 1.3, hr.V100(), 1e-9)
		require.InDelta(t, 1.472, hr.V40(), 1e-9)
		require.InDelta(t, 1.3286666666666667, hr.TargetMoistureVoltage(), 1e-9)
	})
}

func TestHumidityReference_LowAndSuspicious(t *testing.T) {
	t.Parallel()

	hr := ml.NewHumidityReference(1.8, 1.2)
	tests := []struct {
		name       string
		voltage    float64
		low        bool
		suspicious bool
	}{
		{name: "with 50 percent, then it is neither low nor suspicious", voltage: 1.7},
		{name: "with 40 percent, then it is low", voltage: 1.8, low: true},
		{name: "with 0 percent, then it is low", voltage: 2.2, low: true},
		{name: "with -30 percent, then it is low but not suspicious", voltage: 2.5, low: true},
		{name: "with -40 percent, then it is low and suspicious", voltage: 2.6, low: true, suspicious: true},
	}
	for _, tt := range tests {
		t.Run(`Given a HumidityReference struct,
		when IsLow and IsSuspicious methods are called `+tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.low, hr.IsLow(tt.voltage))
			require.Equal(t, tt.suspicious, hr.IsSuspicious(tt.voltage))
		})
	}
}
