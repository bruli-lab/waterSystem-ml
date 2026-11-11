package app

import (
	"context"
	"errors"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

const TrainingZoneCommandName = "training_zone"

type TrainingZoneCommand struct {
	Zone  string
	Force bool
}

func (t TrainingZoneCommand) Name() string {
	return TrainingZoneCommandName
}

type TrainingZone struct {
	trainer   TrainExecutor
	stateRepo ModelTrainingStateRepository
	logRepo   ModelTrainingLogRepository
	now       func() time.Time
}

func (t TrainingZone) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(TrainingZoneCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(TrainingZoneCommandName, cmd.Name())
	}
	if !co.Force {
		recently, err := t.recentlyTrained(ctx, co.Zone)
		if err != nil || recently {
			return nil, err
		}
	}
	if err := t.trainer.Run(ctx, co.Zone); err != nil {
		return nil, err
	}
	trainedAt := t.now()
	if err := t.stateRepo.Save(ctx, ml.NewModelTrainingState(co.Zone, trainedAt)); err != nil {
		return nil, err
	}
	return nil, t.logRepo.Save(ctx, co.Zone, trainedAt)
}

func (t TrainingZone) recentlyTrained(ctx context.Context, zone string) (bool, error) {
	mts, err := t.stateRepo.GetModelTrainingState(ctx, zone)
	if errors.Is(err, ml.ErrModelTrainingStateNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return mts.IsRecentlyTraining(t.now()), nil
}

func NewTrainingZone(
	trainer TrainExecutor,
	stateRepo ModelTrainingStateRepository,
	logRepo ModelTrainingLogRepository,
	now func() time.Time,
) *TrainingZone {
	return &TrainingZone{trainer: trainer, stateRepo: stateRepo, logRepo: logRepo, now: now}
}
