package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
)

const TrainDryingModelCommandName = "train_drying_model"

type TrainDryingModelCommand struct {
	Zone      string
	IfMissing bool
}

func (t TrainDryingModelCommand) Name() string {
	return TrainDryingModelCommandName
}

type TrainDryingModel struct {
	trainer DryingTrainExecutor
}

func (t TrainDryingModel) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(TrainDryingModelCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(TrainDryingModelCommandName, cmd.Name())
	}
	return nil, t.trainer.Run(ctx, co.Zone, co.IfMissing)
}

func NewTrainDryingModel(trainer DryingTrainExecutor) *TrainDryingModel {
	return &TrainDryingModel{trainer: trainer}
}
