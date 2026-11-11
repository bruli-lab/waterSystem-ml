package app

import (
	"context"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
)

const SaveWateringSkippedLogCommandName = "save_watering_skipped_log"

type SaveWateringSkippedLogCommand struct {
	Zone           *string
	Reason         string
	Moisture       *float64
	PredictionID   *uuid.UUID
	DecisionReason *string
	WateringProba  *float64
	ExecutedAt     time.Time
}

func (w SaveWateringSkippedLogCommand) Name() string {
	return SaveWateringSkippedLogCommandName
}

type SaveWateringSkippedLog struct {
	repo WateringSkippedLogRepository
}

func (w SaveWateringSkippedLog) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(SaveWateringSkippedLogCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(SaveWateringSkippedLogCommandName, cmd.Name())
	}
	ws := ml.NewWateringSkippedLog(co.Zone, co.Reason, co.Moisture, co.PredictionID, co.DecisionReason, co.WateringProba, co.ExecutedAt)
	return nil, w.repo.Save(ctx, ws)
}

func NewSaveWateringSkippedLog(repo WateringSkippedLogRepository) *SaveWateringSkippedLog {
	return &SaveWateringSkippedLog{repo: repo}
}
