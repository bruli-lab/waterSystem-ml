package app

import (
	"context"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type PredictionLogRepository interface {
	Save(ctx context.Context, pl *ml.PredictionLog) error
	GetPendingByZone(ctx context.Context, zone string, limit time.Time) (*ml.PredictionLog, error)
	GetPendingValidationZones(ctx context.Context) (map[string]bool, error)
}
