package postgres

import (
	"context"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type PredictionLogRepository struct {
	db     bun.IDB
	tracer trace.Tracer
}

func (p PredictionLogRepository) Save(ctx context.Context, pl *ml.PredictionLog) error {
	ctx, span := p.tracer.Start(ctx, "PredictionLogRepository.Save")
	defer span.End()
	model := buildModel(pl)
	_, err := p.db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("validation_at = EXCLUDED.validation_at").
		Set("moisture_after = EXCLUDED.moisture_after").
		Set("reached_target = EXCLUDED.reached_target").
		Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func buildModel(pl *ml.PredictionLog) *modelPrediction {
	return &modelPrediction{
		BaseModel:        bun.BaseModel{},
		ID:               pl.Id(),
		CreatedAt:        pl.CreatedAt(),
		Zone:             pl.Zone(),
		ShouldWater:      pl.ShouldWater(),
		PredictedSeconds: pl.PredictedSeconds(),
		DecisionReason:   pl.DecisionReason(),
		MoistureBefore:   pl.MoistureBefore(),
		WateringExecuted: pl.WateringExecuted(),
		TargetMoisture:   pl.TargetMoisture(),
		ValidationStatus: pl.Status().String(),
		ValidationAt:     pl.ValidationAt(),
		MoistureAfter:    pl.MoistureAfter(),
		ReachedTarget:    pl.ReachedTarget(),
	}
}

func NewPredictionLogRepository(db bun.IDB, tracer trace.Tracer) *PredictionLogRepository {
	return &PredictionLogRepository{db: db, tracer: tracer}
}
