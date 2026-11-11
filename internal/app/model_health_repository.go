package app

import (
	"context"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ModelHealthRepository interface {
	GetModelHealth(ctx context.Context, zone string, since time.Time) (*ml.ModelHealth, error)
}
