package influxdb2

import (
	"context"
	"fmt"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ForecastRepository struct {
	client influxdb.Client
	org    string
	bucket string
	tracer trace.Tracer
}

func (f ForecastRepository) TemperatureAt(ctx context.Context, at time.Time) (*float64, error) {
	ctx, span := f.tracer.Start(ctx, "ForecastRepository.TemperatureAt")
	defer span.End()

	query := fmt.Sprintf(`
from(bucket: %q)
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "forecast_v2")
  |> filter(fn: (r) => r._field == "temperature")
  |> group()
  |> last()
`, f.bucket, at.Add(-2*time.Hour).UTC().Format(time.RFC3339), at.Add(time.Second).UTC().Format(time.RFC3339))

	result, err := f.client.QueryAPI(f.org).Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error querying forecast temperature: %w", err)
	}

	var temperature *float64
	for result.Next() {
		if value, ok := result.Record().Value().(float64); ok {
			temperature = &value
		}
	}
	if result.Err() != nil {
		err := fmt.Errorf("error reading forecast temperature: %w", result.Err())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "OK")
	return temperature, nil
}

func NewForecastRepository(url, token, org, bucket string, tracer trace.Tracer) *ForecastRepository {
	return &ForecastRepository{client: influxdb.NewClient(url, token), org: org, bucket: bucket, tracer: tracer}
}
