package app_test

import (
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	bonsaiBigZone   = "bonsai big"
	bonsaiSmallZone = "bonsai small"
)

var fixedNow = time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

func clock() time.Time {
	return fixedNow
}

func buildTracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer("test")
}

const (
	humidity100    = 1.2
	humidity40     = 1.8
	lowHumidity    = 1.85
	mediumHumidity = 1.7
)

func dayTimeFunc() time.Time {
	return time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC)
}
