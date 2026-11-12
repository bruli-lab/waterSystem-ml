package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bruli-lab/go-core/ptr"
	"github.com/stretchr/testify/require"

	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/fixtures"
)

func TestValidatePrediction_Handle(t *testing.T) {
	t.Parallel()

	errTest := errors.New("error")
	measures := []ml.SoilMeasure{
		ptr.FromPointer(ml.NewSoilMeasure(bonsaiBigZone, 20)),
		ptr.FromPointer(ml.NewSoilMeasure(bonsaiSmallZone, 20)),
	}
	tests := []struct {
		name                                string
		measures                            []ml.SoilMeasure
		soilMeasureErr, pendingErr, saveErr error
		pendingZones                        map[string]bool
		expectedErr                         error
		expectedSaves, expectedEvents       int
		stale                               bool
		staleErr                            error
	}{
		{
			name:           "and soil measure returns an error, then it returns same error",
			soilMeasureErr: errTest,
			expectedErr:    errTest,
		},
		{
			name:        "and get pending prediction returns an error, then it returns same error",
			measures:    measures,
			pendingErr:  errTest,
			expectedErr: errTest,
		},
		{
			name:         "and save returns an error, then it returns same error",
			measures:     measures,
			pendingZones: map[string]bool{bonsaiBigZone: true},
			saveErr:      errTest,
			expectedErr:  errTest,
		},
		{
			name:        "and getting stale predictions fails, then it returns same error",
			measures:    measures,
			staleErr:    errTest,
			expectedErr: errTest,
		},
		{
			name:          "and a prediction is pending for too long, then it is expired",
			measures:      measures,
			stale:         true,
			expectedSaves: 1,
		},
		{
			name:     "and no zone has a pending prediction, then nothing is saved",
			measures: measures,
		},
		{
			name:           "and only one zone has a pending prediction, then only that one is validated",
			measures:       measures,
			pendingZones:   map[string]bool{bonsaiBigZone: true},
			expectedSaves:  1,
			expectedEvents: 1,
		},
		{
			name:           "and every zone has a pending prediction, then all of them are validated",
			measures:       measures,
			pendingZones:   map[string]bool{bonsaiBigZone: true, bonsaiSmallZone: true},
			expectedSaves:  2,
			expectedEvents: 2,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a ValidatePrediction handler,
		when Handle method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			soilMeasureRepo := &app.SoilMeasureRepositoryMock{
				GetFunc: func(_ context.Context) ([]ml.SoilMeasure, error) {
					return tt.measures, tt.soilMeasureErr
				},
			}
			predictionLogRepo := &app.PredictionLogRepositoryMock{
				GetPendingByZoneFunc: func(_ context.Context, zone string, _ time.Time) (*ml.PredictionLog, error) {
					if tt.pendingErr != nil {
						return nil, tt.pendingErr
					}
					if !tt.pendingZones[zone] {
						return nil, ml.ErrPredictionLogNotFound
					}
					return new(fixtures.PredictionLogBuilder{Zone: &zone}.Build(t)), nil
				},
				SaveFunc: func(_ context.Context, _ *ml.PredictionLog) error {
					return tt.saveErr
				},
				GetPendingBeforeFunc: func(_ context.Context, before time.Time) ([]*ml.PredictionLog, error) {
					require.Equal(t, fixedNow.Add(-ml.MaxPendingValidationAge), before)
					if !tt.stale {
						return nil, tt.staleErr
					}
					zone := bonsaiSmallZone
					return []*ml.PredictionLog{new(fixtures.PredictionLogBuilder{Zone: &zone}.Build(t))}, tt.staleErr
				},
			}
			handler := app.NewValidatePrediction(soilMeasureRepo, predictionLogRepo, clock)

			events, err := handler.Handle(t.Context(), app.ValidatePredictionCommand{Limit: fixedNow})
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, events, tt.expectedEvents)
			saves := predictionLogRepo.SaveCalls()
			require.Len(t, saves, tt.expectedSaves)
			for _, call := range saves {
				require.Equal(t, fixedNow, *call.Pl.ValidationAt())
			}
			if tt.stale {
				last := saves[len(saves)-1].Pl
				require.Equal(t, ml.PredictionLogStatusExpired, last.Status())
				require.Nil(t, last.MoistureAfter())
			}
		})
	}
}
