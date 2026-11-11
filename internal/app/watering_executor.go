package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/watering"
)

type WateringExecutor interface {
	Execute(ctx context.Context, w *watering.Watering) error
}
