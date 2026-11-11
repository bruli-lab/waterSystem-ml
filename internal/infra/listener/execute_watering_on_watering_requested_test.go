package listener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/infra/listener"
)

type wateringCommandRecorder struct {
	executeErr   error
	publishCalls int
	savedLogs    []app.SavePredictionLogCommand
}

func (w *wateringCommandRecorder) Handle(_ context.Context, cmd cqs.Command) ([]event.Event, error) {
	switch c := cmd.(type) {
	case app.ExecuteWateringCommand:
		return nil, w.executeErr
	case app.PublishMessageCommand:
		w.publishCalls++
	case app.SavePredictionLogCommand:
		w.savedLogs = append(w.savedLogs, c)
	}
	return nil, nil
}

func TestExecuteWateringOnWateringRequested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		executeErr       error
		expectedPublish  int
		expectedExecuted bool
	}{
		{
			name:             "and watering succeeds, then the prediction log is saved as executed",
			expectedExecuted: true,
		},
		{
			name:            "and watering fails, then it notifies, saves the log as not executed and does not return an error",
			executeErr:      errors.New("boom"),
			expectedPublish: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &wateringCommandRecorder{executeErr: tt.executeErr}
			l := listener.NewExecuteWateringOnWateringRequested(rec)
			predictionID := uuid.New()
			reason := "reason"
			ev := ml.NewWateringRequestedEvent(zone, ml.ModelPredictionReason, 20, 1.7, 1.3, &predictionID, &reason, nil)

			require.NoError(t, l.Listen(t.Context(), ev))
			require.Equal(t, tt.expectedPublish, rec.publishCalls)
			require.Len(t, rec.savedLogs, 1)
			require.Equal(t, tt.expectedExecuted, rec.savedLogs[0].WateringExecuted)
		})
	}
}
