package app

import (
	"context"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/domain/ml"
)

const TrainingZoneCommandName = "training_zone"

type TrainingZoneCommand struct {
	Zone string
}

func (t TrainingZoneCommand) Name() string {
	return TrainingZoneCommandName
}

type TrainingZone struct {
	svc *ml.Train
}

func (t TrainingZone) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(TrainingZoneCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(TrainingZoneCommandName, cmd.Name())
	}
	return nil, t.svc.Run(ctx, co.Zone)
}

func NewTrainingZone(svc *ml.Train) *TrainingZone {
	return &TrainingZone{svc: svc}
}
