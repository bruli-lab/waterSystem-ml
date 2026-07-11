package memory

import (
	"context"
	"fmt"

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

func NewHumidityReferenceRepository(bb100, bb40, bs100, bs40 float64) *HumidityReferenceRepository {
	refMap := map[string]*ml.HumidityReference{
		"Bonsai big":   ml.NewHumidityReference(bb40, bb100),
		"Bonsai small": ml.NewHumidityReference(bs40, bs100),
	}
	return &HumidityReferenceRepository{references: refMap}
}
