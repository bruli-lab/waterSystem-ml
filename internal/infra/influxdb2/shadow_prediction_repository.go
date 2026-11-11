package influxdb2

import (
	"context"
	"fmt"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const shadowPredictionMeasurement = "shadow_prediction"

type ShadowPredictionRepository struct {
	client influxdb.Client
	org    string
	bucket string
	tracer trace.Tracer
}

func (s ShadowPredictionRepository) Save(ctx context.Context, predictions []ml.ShadowPrediction) error {
	ctx, span := s.tracer.Start(ctx, "ShadowPredictionRepository.Save")
	defer span.End()
	points := make([]*write.Point, 0, len(predictions))
	for i := range predictions {
		points = append(points, shadowPoint(&predictions[i]))
	}
	if err := s.client.WriteAPIBlocking(s.org, s.bucket).WritePoint(ctx, points...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("write shadow predictions to influxdb: %w", err)
	}
	span.SetStatus(codes.Ok, "shadow predictions saved")
	return nil
}

func shadowPoint(p *ml.ShadowPrediction) *write.Point {
	tags := map[string]string{"zone": p.Zone(), "status": "ok"}
	fields := map[string]any{"would_water": p.WouldWater()}
	if h := p.Humidity(); h != nil {
		fields["humidity"] = *h
	}
	if p.Failed() {
		tags["status"] = "error"
		fields["error"] = p.Error()
		return write.NewPoint(shadowPredictionMeasurement, tags, fields, p.PredictedAt())
	}
	t := p.Trajectory()
	fields["start_humidity"] = t.StartHumidity()
	fields["horizon_hours"] = t.Horizon().Hours()
	fields["night"] = t.Horizon().Night()
	fields["min_pessimistic"] = t.Minimum().Pessimistic()
	fields["min_median"] = t.Minimum().Median()
	fields["pessimistic_1h"] = t.At1h().Pessimistic()
	fields["median_1h"] = t.At1h().Median()
	fields["pessimistic_3h"] = t.At3h().Pessimistic()
	fields["median_3h"] = t.At3h().Median()
	fields["model_trained_at"] = t.ModelTrainedAt()
	if rain := t.RainProbabilityMax(); rain != nil {
		fields["rain_probability_max"] = *rain
	}
	return write.NewPoint(shadowPredictionMeasurement, tags, fields, p.PredictedAt())
}

func NewShadowPredictionRepository(url, token, org, bucket string, tracer trace.Tracer) *ShadowPredictionRepository {
	return &ShadowPredictionRepository{client: influxdb.NewClient(url, token), org: org, bucket: bucket, tracer: tracer}
}
