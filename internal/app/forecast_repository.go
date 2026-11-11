package app

import (
	"context"
	"time"
)

type ForecastRepository interface {
	TemperatureAt(ctx context.Context, at time.Time) (*float64, error)
}
