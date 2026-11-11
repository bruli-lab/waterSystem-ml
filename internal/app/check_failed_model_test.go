package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

func TestCheckFailedModel_Handle(t *testing.T) {
	errTest := errors.New("error")
	tests := []struct {
		name                                           string
		expectedErr, getTrainingStateErr, getHealthErr error
		modelHealth                                    *ml.ModelHealth
		modelTrainingState                             *ml.ModelTrainingState
		expectedEvents                                 int
		expectedHealthCalls                            int
		expectedSince                                  time.Time
	}{
		{
			name:                "and get training state returns an error, then it returns same error",
			getTrainingStateErr: errTest,
			expectedErr:         errTest,
		},
		{
			name:                "and get model health returns an error, then it returns same error",
			getTrainingStateErr: ml.ErrModelTrainingStateNotFound,
			getHealthErr:        errTest,
			expectedErr:         errTest,
		},
		{
			name:                "and there is no training state and the model is healthy, then it returns no events",
			getTrainingStateErr: ml.ErrModelTrainingStateNotFound,
			modelHealth:         ml.NewModelHealth(bonsaiBigZone, 19, 1),
			expectedHealthCalls: 1,
		},
		{
			name:               "and the model was trained recently, then health is not checked",
			modelTrainingState: ml.NewModelTrainingState(bonsaiBigZone, fixedNow.Add(-time.Hour)),
		},
		{
			name:                "and the model was not trained recently and is degraded, then it returns one event",
			modelTrainingState:  ml.NewModelTrainingState(bonsaiBigZone, fixedNow.Add(-50*time.Hour)),
			modelHealth:         ml.NewModelHealth(bonsaiBigZone, 10, 10),
			expectedEvents:      1,
			expectedHealthCalls: 1,
			expectedSince:       fixedNow.Add(-50 * time.Hour),
		},
	}
	for _, tt := range tests {
		t.Run(`Given a CheckFailedModel handler,
		when Handle method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			healthRepo := &app.ModelHealthRepositoryMock{
				GetModelHealthFunc: func(_ context.Context, _ string, _ time.Time) (*ml.ModelHealth, error) {
					return tt.modelHealth, tt.getHealthErr
				},
			}
			stateRepo := &app.ModelTrainingStateRepositoryMock{
				GetModelTrainingStateFunc: func(_ context.Context, _ string) (*ml.ModelTrainingState, error) {
					return tt.modelTrainingState, tt.getTrainingStateErr
				},
			}
			handler := app.NewCheckFailedModel(healthRepo, stateRepo, clock)

			events, err := handler.Handle(t.Context(), app.CheckFailedModelCommand{Zone: bonsaiBigZone})
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, events, tt.expectedEvents)
			calls := healthRepo.GetModelHealthCalls()
			require.Len(t, calls, tt.expectedHealthCalls)
			for _, call := range calls {
				require.Equal(t, tt.expectedSince, call.Since)
			}
		})
	}
}
