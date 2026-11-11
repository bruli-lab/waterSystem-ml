package app

import (
	"context"
	"errors"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

const ValidatePredictionCommandName = "validate_prediction"

type ValidatePredictionCommand struct {
	Limit time.Time
}

func (v ValidatePredictionCommand) Name() string {
	return ValidatePredictionCommandName
}

type ValidatePrediction struct {
	soilMeasureRepo   SoilMeasureRepository
	predictionLogRepo PredictionLogRepository
	now               func() time.Time
}

func (v ValidatePrediction) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(ValidatePredictionCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(ValidatePredictionCommandName, cmd.Name())
	}
	return v.validate(ctx, co.Limit)
}

func (v ValidatePrediction) validate(ctx context.Context, limit time.Time) ([]event.Event, error) {
	measures, err := v.soilMeasureRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	events := make([]event.Event, 0)
	for _, measure := range measures {
		pl, err := v.predictionLogRepo.GetPendingByZone(ctx, measure.Zone(), limit)
		if err != nil {
			if errors.Is(err, ml.ErrPredictionLogNotFound) {
				continue
			}
			return nil, err
		}
		pl.AddValidation(new(v.now()), new(measure.Humidity()))
		if err := v.predictionLogRepo.Save(ctx, pl); err != nil {
			return nil, err
		}
		events = append(events, pl.Events()...)
	}
	if err := v.expireStale(ctx, limit); err != nil {
		return nil, err
	}
	return events, nil
}

func (v ValidatePrediction) expireStale(ctx context.Context, limit time.Time) error {
	stale, err := v.predictionLogRepo.GetPendingBefore(ctx, limit.Add(-ml.MaxPendingValidationAge))
	if err != nil {
		return err
	}
	for _, pl := range stale {
		pl.Expire(v.now())
		if err := v.predictionLogRepo.Save(ctx, pl); err != nil {
			return err
		}
	}
	return nil
}

func NewValidatePrediction(
	soilMeasureRepo SoilMeasureRepository,
	predictionLogRepo PredictionLogRepository,
	now func() time.Time,
) *ValidatePrediction {
	return &ValidatePrediction{
		soilMeasureRepo:   soilMeasureRepo,
		predictionLogRepo: predictionLogRepo,
		now:               now,
	}
}
