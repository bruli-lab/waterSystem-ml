package fixtures

import (
	"testing"
	"time"

	"github.com/bruli-lab/go-core/fixtures"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type PredictionLogBuilder struct {
	ID               *uuid.UUID
	Zone             *string
	ShouldWater      bool
	PredictedSeconds *float64
	DecisionReason   *string
	MoistureBefore   *float64
	WateringExecuted bool
	TargetMoisture   *float64
	CreatedAt        *time.Time
}

func (b PredictionLogBuilder) Build(t *testing.T) ml.PredictionLog {
	id := fixtures.SetData(uuid.New(), b.ID)
	zone := fixtures.SetData(uuid.NewString(), b.Zone)
	seconds := fixtures.SetData(float64(10), b.PredictedSeconds)
	reason := fixtures.SetData("reason", b.DecisionReason)
	moisture := fixtures.SetData(float64(10), b.MoistureBefore)
	target := fixtures.SetData(float64(10), b.TargetMoisture)
	createdAt := fixtures.SetData(time.Now(), b.CreatedAt)
	pl, err := ml.NewPredictionLog(id, zone, b.ShouldWater, seconds, reason, moisture, b.WateringExecuted, target, createdAt)
	require.NoError(t, err)
	return *pl
}
