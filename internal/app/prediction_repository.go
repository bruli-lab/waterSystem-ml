package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type PredictionRepository interface {
	Get(ctx context.Context) ([]ml.Prediction, error)
}
