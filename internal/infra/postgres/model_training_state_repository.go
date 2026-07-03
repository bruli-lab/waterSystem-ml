package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type modelTrainingStateRepository struct {
	db     bun.IDB
	tracer trace.Tracer
}

func (m modelTrainingStateRepository) GetModelTrainingState(ctx context.Context, zone string) (*ml.ModelTrainingState, error) {
	ctx, span := m.tracer.Start(ctx, "ModelTrainingStateRepository.GetModelTrainingState")
	defer span.End()
	model := &modelTrainingState{}
	err := m.db.NewSelect().
		Model(model).
		Where("zone = ?", zone).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ml.ErrModelPredictionStateNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return ml.NewModelTrainingState(model.Zone, model.LastTrainingAt), nil
}

func (m modelTrainingStateRepository) Save(ctx context.Context, state *ml.ModelTrainingState) error {
	ctx, span := m.tracer.Start(ctx, "ModelTrainingStateRepository.Save")
	defer span.End()
	model := &modelTrainingState{
		BaseModel:      bun.BaseModel{},
		Zone:           state.Zone(),
		LastTrainingAt: state.TrainingAt(),
	}
	_, err := m.db.NewInsert().Model(model).
		On("CONFLICT (zone) DO UPDATE").
		Set("last_training_at = EXCLUDED.last_training_at").
		Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func NewModelTrainingStateRepository(db bun.IDB, tracer trace.Tracer) *modelTrainingStateRepository {
	return &modelTrainingStateRepository{db: db, tracer: tracer}
}
