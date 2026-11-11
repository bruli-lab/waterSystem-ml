package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type SoilMeasureRepository interface {
	Get(ctx context.Context) ([]ml.SoilMeasure, error)
}
