package ml_test

import (
	"testing"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestNewHumidityReference(t *testing.T) {
	hr := ml.NewHumidityReference(1.472, 1.3)
	require.Equal(t, 1.3, hr.V100())
	require.Equal(t, 1.472, hr.V40())
	require.Equal(t, 1.3286666666666667, hr.TargetMoistureVoltage())
}
