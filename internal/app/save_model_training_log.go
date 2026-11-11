package app

import (
	"context"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
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
	repo ModelTrainingLogRepository
}

func (s SaveModelTrainingLog) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(SaveModelTrainingLogCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(SaveModelTrainingLogCommandName, cmd.Name())
	}
	return nil, s.repo.Save(ctx, co.Zone, co.ExecutedAt)
}

func NewSaveModelTrainingLog(repo ModelTrainingLogRepository) *SaveModelTrainingLog {
	return &SaveModelTrainingLog{repo: repo}
}
