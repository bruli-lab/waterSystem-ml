package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
)

const SavePredictionLogCommandName = "save_prediction_log"

type SavePredictionLogCommand struct {
	ID               uuid.UUID
	Zone             string
	ShouldWater      bool
	PredictedSeconds float64
	DecisionReason   string
	MoistureBefore   float64
	WateringExecuted bool
	TargetMoisture   float64
}

func (s *SavePredictionLogCommand) Name() string {
	return SavePredictionLogCommandName
}

type SavePredictionLog struct {
	svc *ml.SavePredictionLog
}

func (s SavePredictionLog) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(*SavePredictionLogCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(SavePredictionLogCommandName, cmd.Name())
	}
	pl, err := ml.NewPredictionLog(
		co.ID,
		co.Zone,
		co.ShouldWater,
		co.PredictedSeconds,
		co.DecisionReason,
		co.MoistureBefore,
		co.WateringExecuted,
		co.TargetMoisture,
	)
	if err != nil {
		return nil, err
	}
	return nil, s.svc.Save(ctx, pl)
}

func NewSavePredictionLog(svc *ml.SavePredictionLog) *SavePredictionLog {
	return &SavePredictionLog{svc: svc}
}
