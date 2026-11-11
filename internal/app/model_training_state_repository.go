package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ModelTrainingStateRepository interface {
	GetModelTrainingState(ctx context.Context, zone string) (*ml.ModelTrainingState, error)
	Save(ctx context.Context, state *ml.ModelTrainingState) error
}
