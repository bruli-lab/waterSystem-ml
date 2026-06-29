package influxdb2

import (
	"context"
	"fmt"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ModelTrainingLogRepository struct {
	client influxdb.Client
	org    string
	bucket string
	tracer trace.Tracer
}

func (m ModelTrainingLogRepository) Save(ctx context.Context, zone string, at time.Time) error {
	ctx, span := m.tracer.Start(ctx, "ModelTrainingLogRepository.Save")
	defer span.End()
	writeAPI := m.client.WriteAPIBlocking(m.org, m.bucket)

	point := write.NewPoint("model_training", map[string]string{
		"model": zone,
	}, map[string]any{
		"trained": true,
	}, at)
	if err := writeAPI.WritePoint(ctx, point); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("write point to influxdb: %w", err)
	}
	span.SetStatus(codes.Ok, "model training log saved")
	return nil
}

func NewModelTrainingLogRepository(url, token, org, bucket string, tracer trace.Tracer) *ModelTrainingLogRepository {
	client := influxdb.NewClient(url, token)
	return &ModelTrainingLogRepository{client: client, org: org, bucket: bucket, tracer: tracer}
}
