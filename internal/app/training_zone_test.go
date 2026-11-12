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

func TestTrainingZone_Handle(t *testing.T) {
	t.Parallel()

	errTest := errors.New("error")
	tests := []struct {
		name                     string
		force                    bool
		state                    *ml.ModelTrainingState
		getStateErr, trainErr    error
		expectedErr              error
		expectedTrain, expectLog bool
	}{
		{
			name:        "and get training state fails, then it returns same error",
			getStateErr: errTest,
			expectedErr: errTest,
		},
		{
			name:  "and the zone was trained recently, then it does not train",
			state: ml.NewModelTrainingState(bonsaiBigZone, fixedNow.Add(-time.Hour)),
		},
		{
			name:          "and the zone was trained recently but training is forced, then it trains",
			force:         true,
			state:         ml.NewModelTrainingState(bonsaiBigZone, fixedNow.Add(-time.Hour)),
			expectedTrain: true,
			expectLog:     true,
		},
		{
			name:          "and there is no training state, then it trains and saves state and log",
			getStateErr:   ml.ErrModelTrainingStateNotFound,
			expectedTrain: true,
			expectLog:     true,
		},
		{
			name:          "and training fails, then it returns same error and saves nothing",
			getStateErr:   ml.ErrModelTrainingStateNotFound,
			trainErr:      errTest,
			expectedErr:   errTest,
			expectedTrain: true,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a TrainingZone handler,
		when Handle method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			trainer := &app.TrainExecutorMock{
				RunFunc: func(_ context.Context, _ string) error {
					return tt.trainErr
				},
			}
			stateRepo := &app.ModelTrainingStateRepositoryMock{
				GetModelTrainingStateFunc: func(_ context.Context, _ string) (*ml.ModelTrainingState, error) {
					return tt.state, tt.getStateErr
				},
				SaveFunc: func(_ context.Context, _ *ml.ModelTrainingState) error {
					return nil
				},
			}
			logRepo := &app.ModelTrainingLogRepositoryMock{
				SaveFunc: func(_ context.Context, _ string, _ time.Time) error {
					return nil
				},
			}
			handler := app.NewTrainingZone(trainer, stateRepo, logRepo, clock)

			_, err := handler.Handle(t.Context(), app.TrainingZoneCommand{Zone: bonsaiBigZone, Force: tt.force})
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedTrain, len(trainer.RunCalls()) == 1)
			if !tt.expectLog {
				require.Empty(t, stateRepo.SaveCalls())
				require.Empty(t, logRepo.SaveCalls())
				return
			}
			require.Len(t, stateRepo.SaveCalls(), 1)
			require.Equal(t, fixedNow, stateRepo.SaveCalls()[0].State.TrainingAt())
			require.Len(t, logRepo.SaveCalls(), 1)
			require.Equal(t, fixedNow, logRepo.SaveCalls()[0].At)
		})
	}
}
