package listener_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/infra/listener"
	"github.com/stretchr/testify/require"
)

type commandRecorder struct {
	messages []string
}

func (c *commandRecorder) Handle(_ context.Context, cmd cqs.Command) ([]event.Event, error) {
	if m, ok := cmd.(app.PublishMessageCommand); ok {
		c.messages = append(c.messages, m.Message)
	}
	return nil, nil
}

const zone = "bonsai big"

func failed() event.Event {
	return ml.NewWateringZoneSkippedEvent(zone, ml.PredictionErrorReason, 1.7, nil, new("boom"), nil)
}

func TestPublishMessageOnPredictionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		events        []event.Event
		expectedTexts []string
	}{
		{
			name:          "first failure notifies",
			events:        []event.Event{failed()},
			expectedTexts: []string{"ERROR"},
		},
		{
			name:          "repeated failures notify once",
			events:        []event.Event{failed(), failed(), failed()},
			expectedTexts: []string{"ERROR"},
		},
		{
			name: "recovery by no-watering prediction notifies and re-arms",
			events: []event.Event{
				failed(),
				ml.NewWateringZoneSkippedEvent(zone, ml.ModelNotEstimatedReason, 1.7, nil, nil, nil),
				failed(),
			},
			expectedTexts: []string{"ERROR", "RECOVERED", "ERROR"},
		},
		{
			name: "recovery by watering prediction notifies",
			events: []event.Event{
				failed(),
				ml.NewWateringRequestedEvent(zone, ml.ModelPredictionReason, 20, 1.7, 1.3, nil, nil, nil),
			},
			expectedTexts: []string{"ERROR", "RECOVERED"},
		},
		{
			name: "humidity rule decisions are not a recovery",
			events: []event.Event{
				failed(),
				ml.NewWateringZoneSkippedEvent(zone, ml.AboveMaxThresholdReason, 1.45, nil, nil, nil),
				ml.NewWateringRequestedEvent(zone, ml.BelowMinThresholdReason, 20, 1.85, 1.3, nil, nil, nil),
				failed(),
			},
			expectedTexts: []string{"ERROR"},
		},
		{
			name: "non error skips and predictions without previous failure do not notify",
			events: []event.Event{
				ml.NewWateringZoneSkippedEvent(zone, ml.AboveMaxThresholdReason, 1.45, nil, nil, nil),
				ml.NewWateringZoneSkippedEvent(zone, ml.ModelNotEstimatedReason, 1.7, nil, nil, nil),
				ml.NewWateringRequestedEvent(zone, ml.ModelPredictionReason, 20, 1.7, 1.3, nil, nil, nil),
			},
		},
		{
			name: "zones are tracked independently",
			events: []event.Event{
				failed(),
				ml.NewWateringZoneSkippedEvent("bonsai small", ml.PredictionErrorReason, 1.7, nil, new("boom"), nil),
			},
			expectedTexts: []string{"ERROR", "ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := &commandRecorder{}
			l := listener.NewPublishMessageOnPredictionError(rec)
			for _, ev := range tt.events {
				require.NoError(t, l.Listen(context.Background(), ev))
			}

			require.Len(t, rec.messages, len(tt.expectedTexts))
			for i, text := range tt.expectedTexts {
				require.True(t, strings.Contains(rec.messages[i], text), "message %d: %q", i, rec.messages[i])
			}
		})
	}
}
