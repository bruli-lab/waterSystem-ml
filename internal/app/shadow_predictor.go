package app

import (
	"context"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ShadowPredictor interface {
	Predict(ctx context.Context, now time.Time) ([]ml.ShadowPrediction, error)
}
