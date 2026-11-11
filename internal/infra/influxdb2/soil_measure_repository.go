package influxdb2

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/bruli-lab/go-core/ptr"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SoilMeasureRepository struct {
	client       influxdb.Client
	measurements map[string]string
	org          string
	bucket       string
	tracer       trace.Tracer
}

func (s SoilMeasureRepository) Get(ctx context.Context) ([]ml.SoilMeasure, error) {
	ctx, span := s.tracer.Start(ctx, "SoilMeasureRepository.Get")
	defer span.End()

	query := fmt.Sprintf(`
measurements = %s

from(bucket: %q)
  |> range(start: -40m)
  |> filter(fn: (r) => contains(set: measurements, value: r._measurement))
  |> filter(fn: (r) => r._field == "value")
  |> filter(fn: (r) => r._value >= 0.5 and r._value <= 3.3)
  |> group(columns: ["_measurement"])
  |> median()
`, fluxStringArray(slices.Sorted(maps.Keys(s.measurements))), s.bucket)

	result, err := s.client.QueryAPI(s.org).Query(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("error querying soil moisture in influxdb: %w", err)
	}

	measures := make([]ml.SoilMeasure, 0)

	for result.Next() {
		record := result.Record()

		measurement, ok := record.ValueByKey("_measurement").(string)
		if !ok {
			err := errors.New("error parsing soil moisture measurement")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, err
		}

		zone, ok := s.measurements[measurement]
		if !ok {
			err := fmt.Errorf("invalid zone: %s", measurement)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, err
		}

		humidity, ok := record.Value().(float64)
		if !ok {
			err := errors.New("error parsing soil moisture value")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, err
		}

		measures = append(measures, ptr.FromPointer(ml.NewSoilMeasure(zone, humidity)))
	}

	if result.Err() != nil {
		err := fmt.Errorf("error reading soil moisture result: %w", result.Err())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	span.SetStatus(codes.Ok, "OK")
	return measures, nil
}

func NewSoilMeasureRepository(
	url, token, org, bucket string, measurements map[string]string, tracer trace.Tracer,
) *SoilMeasureRepository {
	client := influxdb.NewClient(url, token)
	return &SoilMeasureRepository{
		client: client, measurements: maps.Clone(measurements), org: org, bucket: bucket, tracer: tracer,
	}
}
