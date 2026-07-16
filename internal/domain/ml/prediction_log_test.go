package ml_test

import (
	"testing"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewPredictionLog(t *testing.T) {
	type args struct {
		id               uuid.UUID
		zone             string
		shouldWater      bool
		predictedSeconds float64
		decisionReason   string
		moistureBefore   float64
		wateringExecuted bool
		targetMoisture   float64
	}
	tests := []struct {
		name        string
		args        args
		expectedErr error
	}{
		{
			name: "with an invalid id, then it returns an invalid prediction log id error",
			args: args{
				id: uuid.Nil,
			},
			expectedErr: ml.ErrInvalidPredictionLogID,
		},
		{
			name: "with an invalid zone, then it returns an invalid prediction log zone error",
			args: args{
				id: uuid.New(),
			},
			expectedErr: ml.ErrInvalidPredictionLogZone,
		},
		{
			name: "with an invalid decision reason, then it returns an invalid prediction log decision reason error",
			args: args{
				id:               uuid.New(),
				zone:             "zone",
				shouldWater:      true,
				predictedSeconds: 10,
			},
			expectedErr: ml.ErrInvalidPredictionLogDecisionReason,
		},
		{
			name: "with an invalid moisture before, then it returns an invalid prediction log moisture before error",
			args: args{
				id:               uuid.New(),
				zone:             "zone",
				shouldWater:      true,
				predictedSeconds: 10,
				decisionReason:   "reason",
			},
			expectedErr: ml.ErrInvalidPredictionLogMoistureBefore,
		},
		{
			name: "with an invalid target moisture, then it returns an invalid prediction log target moisture error",
			args: args{
				id:               uuid.New(),
				zone:             "zone",
				shouldWater:      true,
				predictedSeconds: 10,
				decisionReason:   "reason",
				moistureBefore:   60,
				wateringExecuted: true,
			},
			expectedErr: ml.ErrInvalidPredictionLogTargetMoisture,
		},
		{
			name: "with all values, then it returns an valid struct and event",
			args: args{
				id:               uuid.New(),
				zone:             "zone",
				shouldWater:      true,
				predictedSeconds: 10,
				decisionReason:   "reason",
				moistureBefore:   60,
				wateringExecuted: true,
				targetMoisture:   90,
			},
		},
	}
	for _, tt := range tests {
		t.Run(`Given a PredictionLog struct,
		when th constructor is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ml.NewPredictionLog(
				tt.args.id,
				tt.args.zone,
				tt.args.shouldWater,
				tt.args.predictedSeconds,
				tt.args.decisionReason,
				tt.args.moistureBefore,
				tt.args.wateringExecuted,
				tt.args.targetMoisture,
			)
			if err != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.Equal(t, tt.args.id, got.Id())
			require.False(t, got.CreatedAt().IsZero())
			require.True(t, time.Now().Before(got.ValidateAfter()))
			require.Equal(t, tt.args.zone, got.Zone())
			require.Equal(t, tt.args.shouldWater, got.ShouldWater())
			require.Equal(t, tt.args.predictedSeconds, got.PredictedSeconds())
			require.Equal(t, tt.args.decisionReason, got.DecisionReason())
			require.Equal(t, tt.args.moistureBefore, got.MoistureBefore())
			require.Equal(t, tt.args.wateringExecuted, got.WateringExecuted())
			require.Equal(t, tt.args.targetMoisture, got.TargetMoisture())
			require.Equal(t, ml.PredictionLogStatusPending, got.Status())
			require.Nil(t, got.ValidationAt())
			require.Nil(t, got.MoistureAfter())
			require.Nil(t, got.ReachedTarget())
		})
	}
}

func TestPredictionLog_AddValidation(t *testing.T) {
	type args struct {
		at            *time.Time
		moistureAfter *float64
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus ml.PredictionLogStatus
	}{
		{
			name: "with a low moisture after, then it returns a failed status",
			args: args{
				at:            new(time.Now()),
				moistureAfter: new(float64(12)),
			},
			expectedStatus: ml.PredictionLogStatusFailed,
		},
		{
			name: "with a high moisture after, then it returns a success status",
			args: args{
				at:            new(time.Now()),
				moistureAfter: new(float64(8)),
			},
			expectedStatus: ml.PredictionLogStatusSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a built PredictionLog,
		when AddValidation method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			l := fixtures.PredictionLogBuilder{MoistureBefore: new(float64(15))}.Build(t)
			l.AddValidation(tt.args.at, tt.args.moistureAfter)
			require.Equal(t, tt.args.at, l.ValidationAt())
			require.Equal(t, tt.args.moistureAfter, l.MoistureAfter())
			require.Equal(t, tt.expectedStatus, l.Status())
		})
	}
}
