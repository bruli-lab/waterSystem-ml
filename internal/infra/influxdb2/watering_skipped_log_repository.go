package influxdb2

import (
	"context"
	"fmt"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type WateringSkippedLogRepository struct {
	client influxdb.Client
	org    string
	bucket string
	tracer trace.Tracer
}

func (w WateringSkippedLogRepository) Save(ctx context.Context, skp *ml.WateringSkippedLog) error {
	ctx, span := w.tracer.Start(ctx, "WateringSkippedLogRepository.Save")
	defer span.End()
	writeAPI := w.client.WriteAPIBlocking(w.org, w.bucket)

	tags := map[string]string{
		"reason": skp.Reason(),
	}
	if skp.Zone() != nil {
		tags["zone"] = *skp.Zone()
	}
	fields := map[string]any{
		"count": 1,
	}
	if skp.Moisture() != nil {
		fields["moisture"] = *skp.Moisture()
	}
	if skp.WateringProba() != nil {
		fields["watering_proba"] = *skp.WateringProba()
	}
	point := write.NewPoint("watering_skipped", tags, fields, skp.ExecutedAt())
	if err := writeAPI.WritePoint(ctx, point); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("write point to influxdb: %w", err)
	}
	span.SetStatus(codes.Ok, "watering skipped log saved")
	return nil
}

func NewWateringSkippedLogRepository(url, token, org, bucket string, tracer trace.Tracer) *WateringSkippedLogRepository {
	client := influxdb.NewClient(url, token)
	return &WateringSkippedLogRepository{client: client, org: org, bucket: bucket, tracer: tracer}
}
