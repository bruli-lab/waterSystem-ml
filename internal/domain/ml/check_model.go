package ml

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CheckModel struct {
	modelHealthRepo     ModelHealthRepository
	modelTrainStateRepo ModelTrainingStateRepository
	tracer              trace.Tracer
}

func (c CheckModel) Check(ctx context.Context, zone string) (*ModelHealth, error) {
	ctx, span := c.tracer.Start(ctx, "CheckModel.Check")
	defer span.End()
	mts, err := c.modelTrainStateRepo.GetModelTrainingState(ctx, zone)
	if err != nil {
		if !errors.Is(err, ErrModelPredictionStateNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
	}
	if mts != nil && mts.IsRecentlyTraining() {
		return nil, nil
	}
	modelHealth, err := c.modelHealthRepo.GetModelHealth(ctx, zone)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	modelHealth.Check()
	return modelHealth, nil
}

func NewCheckModel(repo ModelHealthRepository, modelTrainingStateRepo ModelTrainingStateRepository, tracer trace.Tracer) *CheckModel {
	return &CheckModel{modelHealthRepo: repo, tracer: tracer, modelTrainStateRepo: modelTrainingStateRepo}
}
