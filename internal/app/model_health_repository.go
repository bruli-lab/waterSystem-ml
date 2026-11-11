package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ModelHealthRepository interface {
	GetModelHealth(ctx context.Context, zone string) (*ml.ModelHealth, error)
}
