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

func TestRunShadowPrediction_Handle(t *testing.T) {
	t.Parallel()

	errTest := errors.New("error")
	prediction := *ml.NewFailedShadowPrediction(bonsaiBigZone, fixedNow, "not_enough_readings", nil)
	tests := []struct {
		name               string
		predictions        []ml.ShadowPrediction
		predictErr, saveEr error
		expectedErr        error
		expectedSave       bool
	}{
		{
			name:        "and the prediction fails, then it returns same error and saves nothing",
			predictErr:  errTest,
			expectedErr: errTest,
		},
		{
			name: "and there are no predictions, then it saves nothing",
		},
		{
			name:         "and there are predictions, then it saves them",
			predictions:  []ml.ShadowPrediction{prediction},
			expectedSave: true,
		},
		{
			name:         "and saving fails, then it returns same error",
			predictions:  []ml.ShadowPrediction{prediction},
			saveEr:       errTest,
			expectedErr:  errTest,
			expectedSave: true,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a RunShadowPrediction handler,
		when Handle method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			predictor := &app.ShadowPredictorMock{
				PredictFunc: func(_ context.Context, now time.Time) ([]ml.ShadowPrediction, error) {
					require.Equal(t, fixedNow, now)
					return tt.predictions, tt.predictErr
				},
			}
			repo := &app.ShadowPredictionRepositoryMock{
				SaveFunc: func(_ context.Context, predictions []ml.ShadowPrediction) error {
					require.Equal(t, tt.predictions, predictions)
					return tt.saveEr
				},
			}
			handler := app.NewRunShadowPrediction(predictor, repo, clock)

			events, err := handler.Handle(t.Context(), app.RunShadowPredictionCommand{})
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Empty(t, events)
			require.Equal(t, tt.expectedSave, len(repo.SaveCalls()) == 1)
		})
	}
}

func TestTrainDryingModel_Handle(t *testing.T) {
	t.Parallel()

	errTest := errors.New("error")
	for name, trainErr := range map[string]error{"succeeds": nil, "fails": errTest} {
		t.Run(`Given a TrainDryingModel handler,
		when Handle method is called and training `+name+`, then it returns the training result`, func(t *testing.T) {
			t.Parallel()
			trainer := &app.DryingTrainExecutorMock{
				RunFunc: func(_ context.Context, zone string, ifMissing bool) error {
					require.Equal(t, bonsaiBigZone, zone)
					require.True(t, ifMissing)
					return trainErr
				},
			}
			cmd := app.TrainDryingModelCommand{Zone: bonsaiBigZone, IfMissing: true}
			_, err := app.NewTrainDryingModel(trainer).Handle(t.Context(), cmd)
			require.ErrorIs(t, err, trainErr)
			require.Len(t, trainer.RunCalls(), 1)
		})
	}
}

func TestShadowHandlers_InvalidCommand(t *testing.T) {
	t.Parallel()

	t.Run(`Given a RunShadowPrediction handler,
		when Handle method is called with another command, then it returns an error`, func(t *testing.T) {
		t.Parallel()
		_, err := app.NewRunShadowPrediction(&app.ShadowPredictorMock{}, &app.ShadowPredictionRepositoryMock{}, clock).
			Handle(t.Context(), app.TrainDryingModelCommand{})
		require.Error(t, err)
	})

	t.Run(`Given a TrainDryingModel handler,
		when Handle method is called with another command, then it returns an error`, func(t *testing.T) {
		t.Parallel()
		_, err := app.NewTrainDryingModel(&app.DryingTrainExecutorMock{}).Handle(t.Context(), app.RunShadowPredictionCommand{})
		require.Error(t, err)
	})
}
