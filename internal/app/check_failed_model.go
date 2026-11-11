package app

import (
	"context"
	"errors"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"

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
}

func (c CheckFailedModel) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(CheckFailedModelCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(CheckFailedModelCommandName, cmd.Name())
	}
	return c.check(ctx, co.Zone)
}

func (c CheckFailedModel) check(ctx context.Context, zone string) ([]event.Event, error) {
	mts, err := c.modelTrainStateRepo.GetModelTrainingState(ctx, zone)
	if err != nil && !errors.Is(err, ml.ErrModelTrainingStateNotFound) {
		return nil, err
	}
	var lastTrainingAt time.Time
	if mts != nil {
		if mts.IsRecentlyTraining(c.now()) {
			return nil, nil
		}
		lastTrainingAt = mts.TrainingAt()
	}
	modelHealth, err := c.modelHealthRepo.GetModelHealth(ctx, zone, lastTrainingAt)
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
) *CheckFailedModel {
	return &CheckFailedModel{
		modelHealthRepo:     modelHealthRepo,
		modelTrainStateRepo: modelTrainStateRepo,
		now:                 now,
	}
}
