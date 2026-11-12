package ml_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

func TestNewShadowHorizon(t *testing.T) {
	t.Parallel()

	t.Run(`Given zero hours,
		when the ShadowHorizon constructor is called, then it returns an invalid shadow horizon error`, func(t *testing.T) {
		t.Parallel()
		_, err := ml.NewShadowHorizon(0, false)
		require.ErrorIs(t, err, ml.ErrInvalidShadowHorizon)
	})

	t.Run(`Given valid hours at night,
		when the ShadowHorizon constructor is called, then it returns the hours and the night flag`, func(t *testing.T) {
		t.Parallel()
		horizon, err := ml.NewShadowHorizon(11, true)
		require.NoError(t, err)
		require.Equal(t, 11, horizon.Hours())
		require.True(t, horizon.Night())
	})
}

func TestShadowPrediction(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)

	t.Run(`Given a ShadowTrajectory,
		when the ShadowPrediction constructor is called, then it returns a prediction with the trajectory`, func(t *testing.T) {
		t.Parallel()
		horizon, err := ml.NewShadowHorizon(3, false)
		require.NoError(t, err)
		rain := 0.4
		trajectory := ml.NewShadowTrajectory(
			55, 56, horizon,
			ml.NewHumidityEstimate(39, 45),
			ml.NewHumidityEstimate(52, 54),
			ml.NewHumidityEstimate(39, 45),
			&rain, "2026-11-11T20:00:00+00:00",
		)

		ok := ml.NewShadowPrediction("zone", now, trajectory, true)
		require.False(t, ok.Failed())
		require.True(t, ok.WouldWater())
		require.InDelta(t, 55, *ok.Humidity(), 1e-9)
		require.InDelta(t, 56, ok.Trajectory().StartHumidity(), 1e-9)
		require.InDelta(t, 39, ok.Trajectory().Minimum().Pessimistic(), 1e-9)
		require.InDelta(t, 54, ok.Trajectory().At1h().Median(), 1e-9)
		require.InDelta(t, 0.4, *ok.Trajectory().RainProbabilityMax(), 1e-9)
		require.Equal(t, "2026-11-11T20:00:00+00:00", ok.Trajectory().ModelTrainedAt())
	})

	t.Run(`Given an error and a humidity,
		when the failed ShadowPrediction constructor is called, then it returns a failed prediction without trajectory`,
		func(t *testing.T) {
			t.Parallel()
			humidity := 70.0
			failed := ml.NewFailedShadowPrediction("zone", now, "not_enough_readings", &humidity)
			require.True(t, failed.Failed())
			require.False(t, failed.WouldWater())
			require.Nil(t, failed.Trajectory())
			require.InDelta(t, 70, *failed.Humidity(), 1e-9)
		})
}
