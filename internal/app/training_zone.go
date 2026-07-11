package app

import (
	"context"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
)

const TrainingZoneCommandName = "training_zone"

type TrainingZoneCommand struct {
	Zone string
}

func (t TrainingZoneCommand) Name() string {
	return TrainingZoneCommandName
}

type TrainingZone struct {
	trainSvc *ml.Train
	saveSvc  *ml.SaveModelTrainingState
}

func (t TrainingZone) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(TrainingZoneCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(TrainingZoneCommandName, cmd.Name())
	}
	if err := t.trainSvc.Run(ctx, co.Zone); err != nil {
		return nil, err
	}
	ts := ml.NewModelTrainingState(co.Zone, time.Now())
	return nil, t.saveSvc.Save(ctx, ts)
}

func NewTrainingZone(trainSvc *ml.Train, saveSvc *ml.SaveModelTrainingState) *TrainingZone {
	return &TrainingZone{trainSvc: trainSvc, saveSvc: saveSvc}
}
