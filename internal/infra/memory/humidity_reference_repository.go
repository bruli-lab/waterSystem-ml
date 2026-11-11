package memory

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type HumidityReferenceRepository struct {
	references map[string]*ml.HumidityReference
}

func (h HumidityReferenceRepository) GetByZone(ctx context.Context, zone string) (*ml.HumidityReference, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	hum, ok := h.references[zone]
	if !ok {
		return nil, fmt.Errorf("humidity for zone %s not found", zone)
	}
	return hum, nil
}

func (h HumidityReferenceRepository) Zones(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return slices.Sorted(maps.Keys(h.references)), nil
}

func NewHumidityReferenceRepository(references map[string]*ml.HumidityReference) *HumidityReferenceRepository {
	return &HumidityReferenceRepository{references: maps.Clone(references)}
}
