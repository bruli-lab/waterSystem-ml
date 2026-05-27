package memory

import (
	"context"
	"fmt"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
)

type HumidityRepository struct {
	humidities map[string]*ml.Humidity
}

func (h HumidityRepository) GetByZone(ctx context.Context, zone string) (*ml.Humidity, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	hum, ok := h.humidities[zone]
	if !ok {
		return nil, fmt.Errorf("humidity for zone %s not found", zone)
	}
	return hum, nil
}

func NewHumidityRepository(bb100, bb40, bs100, bs40 float64) *HumidityRepository {
	humMap := map[string]*ml.Humidity{
		"Bonsai big":   ml.NewHumidity(bb40, bb100),
		"Bonsai small": ml.NewHumidity(bs40, bs100),
	}
	return &HumidityRepository{humidities: humMap}
}
