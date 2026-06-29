package app

import (
	"context"
	"time"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/domain/ml"
)

const SaveModelTrainingLogCommandName = "save_model_training_log"

type SaveModelTrainingLogCommand struct {
	Zone       string
	ExecutedAt time.Time
}

func (s SaveModelTrainingLogCommand) Name() string {
	return SaveModelTrainingLogCommandName
}

type SaveModelTrainingLog struct {
	svc *ml.SaveModelTrainingLog
}

func (s SaveModelTrainingLog) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(SaveModelTrainingLogCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(SaveModelTrainingLogCommandName, cmd.Name())
	}
	return nil, s.svc.Save(ctx, co.Zone, co.ExecutedAt)
}

func NewSaveModelTrainingLog(svc *ml.SaveModelTrainingLog) *SaveModelTrainingLog {
	return &SaveModelTrainingLog{svc: svc}
}
