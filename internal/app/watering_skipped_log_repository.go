package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type WateringSkippedLogRepository interface {
	Save(ctx context.Context, skp *ml.WateringSkippedLog) error
}
