package app

import (
	"context"
	"errors"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

const CheckFailedModelCommandName = "check_failed_model"

type CheckFailedModelCommand struct {
	Zone string
}

func (c CheckFailedModelCommand) Name() string {
	return CheckFailedModelCommandName
}

type CheckFailedModel struct {
	modelHealthRepo     ModelHealthRepository
	modelTrainStateRepo ModelTrainingStateRepository
	now                 func() time.Time
	tracer              trace.Tracer
}

func (c CheckFailedModel) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(CheckFailedModelCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(CheckFailedModelCommandName, cmd.Name())
	}
	ctx, span := c.tracer.Start(ctx, "app.CheckFailedModel")
	defer span.End()

	events, err := c.check(ctx, co.Zone)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return events, nil
}

func (c CheckFailedModel) check(ctx context.Context, zone string) ([]event.Event, error) {
	mts, err := c.modelTrainStateRepo.GetModelTrainingState(ctx, zone)
	if err != nil && !errors.Is(err, ml.ErrModelPredictionStateNotFound) {
		return nil, err
	}
	if mts != nil && mts.IsRecentlyTraining(c.now()) {
		return nil, nil
	}
	modelHealth, err := c.modelHealthRepo.GetModelHealth(ctx, zone)
	if err != nil {
		return nil, err
	}
	modelHealth.Check()
	return modelHealth.Events(), nil
}

func NewCheckFailedModel(
	modelHealthRepo ModelHealthRepository,
	modelTrainStateRepo ModelTrainingStateRepository,
	now func() time.Time,
	tracer trace.Tracer,
) *CheckFailedModel {
	return &CheckFailedModel{
		modelHealthRepo:     modelHealthRepo,
		modelTrainStateRepo: modelTrainStateRepo,
		now:                 now,
		tracer:              tracer,
	}
}
