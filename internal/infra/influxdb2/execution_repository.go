package influxdb2

import (
	"context"
	"fmt"
	"slices"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ExecutionRepository struct {
	client influxdb.Client
	zones  []string
	org    string
	bucket string
	tracer trace.Tracer
}

func (e *ExecutionRepository) GetLastExecution(ctx context.Context) (ml.Executions, error) {
	ctx, span := e.tracer.Start(ctx, "ExecutionRepository.GetLastExecution")
	defer span.End()
	query := fmt.Sprintf(`
zones = %s

from(bucket: %q)
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "logs")
  |> filter(fn: (r) => r._field == "seconds")
  |> filter(fn: (r) => contains(set: zones, value: r.zone))
  |> group(columns: ["zone"])
  |> last()
  |> keep(columns: ["_time", "zone"])
`, fluxStringArray(e.zones), e.bucket)

	result, err := e.client.QueryAPI(e.org).Query(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("error querying last watering: %w", err)
	}

	lastWatering := make(ml.Executions)

	for result.Next() {
		record := result.Record()

		zone, ok := record.ValueByKey("zone").(string)
		if !ok {
			continue
		}
		exec := ml.NewExecution(zone, record.Time())
		lastWatering[zone] = *exec
	}

	if result.Err() != nil {
		err := fmt.Errorf("query error: %w", result.Err())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	span.SetStatus(codes.Ok, "OK")
	return lastWatering, nil
}

func NewExecutionRepository(url, token, org, bucket string, zones []string, tracer trace.Tracer) *ExecutionRepository {
	client := influxdb.NewClient(url, token)
	return &ExecutionRepository{client: client, zones: slices.Clone(zones), org: org, bucket: bucket, tracer: tracer}
}
