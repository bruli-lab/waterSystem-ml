package python

import (
	"testing"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestParseAndBuildShadowPredictions(t *testing.T) {
	t.Parallel()

	output := `[
  {"zone": "Bonsai big", "humidity": 52.3, "start_humidity": 53.1, "horizon_hours": 3, "night": false,
   "min_pessimistic": 38.2, "min_median": 44.0, "pessimistic_1h": 48.0, "pessimistic_3h": 38.2,
   "median_1h": 50.1, "median_3h": 44.0, "rain_probability_max": null, "would_water": true,
   "model_trained_at": "2026-11-11T20:00:00+00:00"},
  {"zone": "Bonsai small", "humidity": 70.0, "error": "not_enough_readings"},
  {"zone": "Other", "error": "drying_model_not_found"},
  {"zone": "Bonsai big", "humidity": 52.3, "horizon_hours": 0}
]`
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)

	parsed, err := parseShadowResults(output)
	require.NoError(t, err)
	preds := buildShadowDomain(parsed, now)
	require.Len(t, preds, 4)

	require.False(t, preds[0].Failed())
	require.True(t, preds[0].WouldWater())
	require.Equal(t, now, preds[0].PredictedAt())
	trajectory := preds[0].Trajectory()
	require.Equal(t, 3, trajectory.Horizon().Hours())
	require.False(t, trajectory.Horizon().Night())
	require.InDelta(t, 38.2, trajectory.Minimum().Pessimistic(), 1e-9)
	require.InDelta(t, 44.0, trajectory.Minimum().Median(), 1e-9)
	require.InDelta(t, 48.0, trajectory.At1h().Pessimistic(), 1e-9)
	require.InDelta(t, 44.0, trajectory.At3h().Median(), 1e-9)
	require.Nil(t, trajectory.RainProbabilityMax())

	require.True(t, preds[1].Failed())
	require.Equal(t, "not_enough_readings", preds[1].Error())
	require.InDelta(t, 70.0, *preds[1].Humidity(), 1e-9)

	require.True(t, preds[2].Failed())
	require.Nil(t, preds[2].Humidity())

	require.True(t, preds[3].Failed())
	require.Equal(t, ml.ErrInvalidShadowHorizon.Error(), preds[3].Error())
}

func TestParseShadowResults_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseShadowResults("not json")
	require.Error(t, err)
}
