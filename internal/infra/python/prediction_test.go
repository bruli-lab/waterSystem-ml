package python

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndBuildPredictions(t *testing.T) {
	t.Parallel()

	t.Run(`Given a predict.py output with a watering zone and a failed zone,
		when it is parsed and built, then it returns a watering prediction and a failed prediction`, func(t *testing.T) {
		t.Parallel()
		output := `[
  {"zone": "Bonsai big", "should_water": true, "decision_reason": "model",
   "predicted_seconds": 20, "probability": 0.72, "used_model": true},
  {"zone": "Bonsai small", "should_water": false, "decision_reason": "Error loading model: boom",
   "predicted_seconds": 0, "probability": 0, "used_model": false, "error": "Error loading model: boom"}
]`

		parsed, err := parsePredictions(output)
		require.NoError(t, err)

		preds := buildPredictionsDomain(parsed)
		require.Len(t, preds, 2)

		require.False(t, preds[0].Failed())
		require.True(t, preds[0].ShouldWater())
		require.InDelta(t, 0.72, preds[0].WateringProba(), 1e-9)
		require.InDelta(t, 20, preds[0].PredictedSeconds(), 1e-9)

		require.True(t, preds[1].Failed())
		require.False(t, preds[1].ShouldWater())
		require.Equal(t, "Error loading model: boom", preds[1].ErrorMessage())
	})
}

func TestTail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		limit    int
		expected string
	}{
		{name: "with a text shorter than the limit, then it returns the whole text", text: "abc", limit: 5, expected: "abc"},
		{name: "with a text longer than the limit, then it returns the last characters", text: "abcde", limit: 3, expected: "cde"},
	}
	for _, tt := range tests {
		t.Run(`Given a text,
		when tail is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tail(tt.text, tt.limit))
		})
	}
}
