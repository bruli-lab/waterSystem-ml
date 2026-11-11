package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type HumidityReferenceRepository interface {
	GetByZone(ctx context.Context, zone string) (*ml.HumidityReference, error)
	Zones(ctx context.Context) ([]string, error)
}
