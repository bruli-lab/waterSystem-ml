package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bruli-lab/go-core/ptr"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCalculateWatering_Handle(t *testing.T) {
	errTest := errors.New("error")

	defaultStatus := ml.NewStatus(true, false)
	defaultExecutions := ml.Executions{
		bonsaiBigZone:   ptr.FromPointer(ml.NewExecution(bonsaiBigZone, dayTimeFunc().Add(-5*time.Hour))),
		bonsaiSmallZone: ptr.FromPointer(ml.NewExecution(bonsaiSmallZone, dayTimeFunc().Add(-5*time.Hour))),
	}
	defaultMeasures := []ml.SoilMeasure{
		ptr.FromPointer(ml.NewSoilMeasure(bonsaiBigZone, mediumHumidity)),
		ptr.FromPointer(ml.NewSoilMeasure(bonsaiSmallZone, mediumHumidity)),
	}
	defaultHumidityRef := ml.NewHumidityReference(humidity40, humidity100)

	tests := []struct {
		name string

		expectedErr error

		statusErr        error
		executionsErr    error
		soilMeasureErr   error
		predictionLogErr error
		humidityRefErr   error
		predictErr       error

		status                *ml.Status
		executions            ml.Executions
		measures              []ml.SoilMeasure
		humRef                *ml.HumidityReference
		pendingValidationZone map[string]bool
		prediction            []ml.Prediction
		wantEvents            int
		wantPredict           bool
	}{
		{
			name:        "and status repository returns error, then it returns same error",
			expectedErr: errTest,
			statusErr:   errTest,
		},
		{
			name:          "and executions repository returns error, then it returns same error",
			expectedErr:   errTest,
			status:        defaultStatus,
			executionsErr: errTest,
		},
		{
			name:           "and soil measure repository returns error, then it returns same error",
			expectedErr:    errTest,
			status:         defaultStatus,
			executions:     defaultExecutions,
			soilMeasureErr: errTest,
		},
		{
			name:             "and prediction log repository returns error, then it returns same error",
			expectedErr:      errTest,
			status:           defaultStatus,
			executions:       defaultExecutions,
			measures:         defaultMeasures,
			predictionLogErr: errTest,
		},
		{
			name:           "and humidity reference repository returns error, then it returns same error",
			expectedErr:    errTest,
			status:         defaultStatus,
			executions:     defaultExecutions,
			measures:       defaultMeasures,
			humidityRefErr: errTest,
		},
		{
			name:        "and calculated watering returns error, then it returns same error",
			expectedErr: ml.ErrUnknownZone,
			status:      defaultStatus,
			executions: ml.Executions{
				"wrong": ptr.FromPointer(ml.NewExecution("wrong", dayTimeFunc().Add(-5*time.Hour))),
			},
			measures: []ml.SoilMeasure{
				ptr.FromPointer(ml.NewSoilMeasure(bonsaiBigZone, lowHumidity)),
				ptr.FromPointer(ml.NewSoilMeasure(bonsaiSmallZone, lowHumidity)),
			},
			humRef: defaultHumidityRef,
		},
		{
			name:        "and prediction repository returns error for medium humidity zones, then it returns same error",
			expectedErr: errTest,
			predictErr:  errTest,
			status:      defaultStatus,
			executions:  defaultExecutions,
			measures:    defaultMeasures,
			humRef:      defaultHumidityRef,
			wantPredict: true,
		},
		{
			name:       "and prediction repository returns predictions for medium humidity zones, then it returns a calculated object",
			status:     defaultStatus,
			executions: defaultExecutions,
			measures:   defaultMeasures,
			humRef:     defaultHumidityRef,
			prediction: []ml.Prediction{
				ptr.FromPointer(ml.NewPrediction(uuid.New(), bonsaiBigZone, false, 0, "no needed", 0)),
				ptr.FromPointer(ml.NewPrediction(uuid.New(), bonsaiSmallZone, false, 0, "no needed", 0)),
			},
			wantEvents:  2,
			wantPredict: true,
		},
		{
			name:       "and one zone has pending validation, then only the other medium zone asks prediction",
			status:     defaultStatus,
			executions: defaultExecutions,
			measures:   defaultMeasures,
			humRef:     defaultHumidityRef,
			pendingValidationZone: map[string]bool{
				bonsaiBigZone: true,
			},
			prediction: []ml.Prediction{
				ptr.FromPointer(ml.NewPrediction(uuid.New(), bonsaiSmallZone, false, 0, "no needed", 0)),
			},
			wantEvents:  2,
			wantPredict: true,
		},
		{
			name:       "and all medium zones have pending validation, then it returns without asking predictions",
			status:     defaultStatus,
			executions: defaultExecutions,
			measures:   defaultMeasures,
			humRef:     defaultHumidityRef,
			pendingValidationZone: map[string]bool{
				bonsaiBigZone:   true,
				bonsaiSmallZone: true,
			},
			wantEvents: 2,
		},
		{
			name:       "and calculated watering is already resolved by low humidity, then it returns without asking predictions",
			status:     defaultStatus,
			executions: defaultExecutions,
			measures: []ml.SoilMeasure{
				ptr.FromPointer(ml.NewSoilMeasure(bonsaiBigZone, lowHumidity)),
				ptr.FromPointer(ml.NewSoilMeasure(bonsaiSmallZone, lowHumidity)),
			},
			humRef:     defaultHumidityRef,
			wantEvents: 2,
		},
		{
			name:       "and system is blocked by raining, then it returns without asking predictions",
			status:     ml.NewStatus(true, true),
			executions: defaultExecutions,
			measures:   defaultMeasures,
			humRef:     defaultHumidityRef,
			wantEvents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(`Given a CalculateWatering handler, when Handle method is called `+tt.name, func(t *testing.T) {
			t.Parallel()

			predictionCalled := false
			predictionRepo := &app.PredictionRepositoryMock{
				GetFunc: func(_ context.Context) ([]ml.Prediction, error) {
					predictionCalled = true
					return tt.prediction, tt.predictErr
				},
			}
			measureRepo := &app.SoilMeasureRepositoryMock{
				GetFunc: func(_ context.Context) ([]ml.SoilMeasure, error) {
					return tt.measures, tt.soilMeasureErr
				},
			}
			humRefRepo := &app.HumidityReferenceRepositoryMock{
				GetByZoneFunc: func(_ context.Context, _ string) (*ml.HumidityReference, error) {
					return tt.humRef, tt.humidityRefErr
				},
			}
			predictionLogRepo := &app.PredictionLogRepositoryMock{
				GetPendingValidationZonesFunc: func(_ context.Context) (map[string]bool, error) {
					return tt.pendingValidationZone, tt.predictionLogErr
				},
			}
			execRepo := &app.ExecutionRepositoryMock{
				GetLastExecutionFunc: func(_ context.Context) (ml.Executions, error) {
					return tt.executions, tt.executionsErr
				},
			}
			statusRepo := &app.StatusRepositoryMock{
				GetStatusFunc: func(_ context.Context) (*ml.Status, error) {
					return tt.status, tt.statusErr
				},
			}

			handler := app.NewCalculateWatering(
				predictionRepo,
				measureRepo,
				humRefRepo,
				execRepo,
				predictionLogRepo,
				statusRepo,
				dayTimeFunc,
				buildTracer(),
			)
			events, err := handler.Handle(t.Context(), app.CalculateWateringCommand{})

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPredict, predictionCalled)
			require.Len(t, events, tt.wantEvents)
		})
	}
}
