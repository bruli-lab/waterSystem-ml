package app

import (
	"context"
	"time"
)

type ModelTrainingLogRepository interface {
	Save(ctx context.Context, zone string, at time.Time) error
}
