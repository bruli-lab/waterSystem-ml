package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ShadowPredictionRepository interface {
	Save(ctx context.Context, predictions []ml.ShadowPrediction) error
}
