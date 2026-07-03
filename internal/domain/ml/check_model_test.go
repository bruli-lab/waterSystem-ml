package ml_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestCheckModel_Check(t *testing.T) {
	errTest := errors.New("error")
	type args struct {
		zone string
	}
	tests := []struct {
		name string
		args args
		expectedErr, getTrainingStateErr,
		getModelHealthErr error
		expectedEventLen   int
		expectedNil        bool
		modelHealth        *ml.ModelHealth
		modelTrainingState *ml.ModelTrainingState
	}{
		{
			name:                "and get training model state returns an error, then it results same error",
			expectedErr:         errTest,
			getTrainingStateErr: errTest,
		},
		{
			name:                "and get model health returns an error, then it results same error",
			expectedErr:         errTest,
			getTrainingStateErr: ml.ErrModelPredictionStateNotFound,
			getModelHealthErr:   errTest,
		},
		{
			name:                "and get model health returns a health model, then it results nil",
			getTrainingStateErr: ml.ErrModelPredictionStateNotFound,
			modelHealth:         ml.NewModelHealth(bonsaiBigZone, 19, 1),
			expectedEventLen:    0,
		},
		{
			name:               "and model training state is executed, then it results no event",
			modelTrainingState: ml.NewModelTrainingState(bonsaiBigZone, time.Now().Add(-1*time.Hour)),
			expectedNil:        true,
		},
		{
			name:               "and model training state is not executed and degraded model, then it results one event",
			modelTrainingState: ml.NewModelTrainingState(bonsaiBigZone, time.Now().Add(-50*time.Hour)),
			modelHealth:        ml.NewModelHealth(bonsaiBigZone, 10, 10),
			expectedEventLen:   1,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a CheckModel service,
		when Check method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			mhRepo := &ml.ModelHealthRepositoryMock{
				GetModelHealthFunc: func(_ context.Context, _ string) (*ml.ModelHealth, error) {
					return tt.modelHealth, tt.getModelHealthErr
				},
			}
			mtsRepo := &ml.ModelTrainingStateRepositoryMock{
				GetModelTrainingStateFunc: func(_ context.Context, _ string) (*ml.ModelTrainingState, error) {
					return tt.modelTrainingState, tt.getTrainingStateErr
				},
			}
			svc := ml.NewCheckModel(mhRepo, mtsRepo, buildTracer())
			got, err := svc.Check(t.Context(), tt.args.zone)
			if err != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			if tt.expectedNil {
				require.Nil(t, got)
				return
			}
			require.Len(t, got.Events(), tt.expectedEventLen)
		})
	}
}
