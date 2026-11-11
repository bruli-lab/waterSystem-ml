package app

import (
	"context"
	"errors"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	tracer            trace.Tracer
}

func (v ValidatePrediction) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(ValidatePredictionCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(ValidatePredictionCommandName, cmd.Name())
	}
	ctx, span := v.tracer.Start(ctx, "app.ValidatePrediction")
	defer span.End()

	events, err := v.validate(ctx, co.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return events, nil
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
	return events, nil
}

func NewValidatePrediction(
	soilMeasureRepo SoilMeasureRepository,
	predictionLogRepo PredictionLogRepository,
	now func() time.Time,
	tracer trace.Tracer,
) *ValidatePrediction {
	return &ValidatePrediction{
		soilMeasureRepo:   soilMeasureRepo,
		predictionLogRepo: predictionLogRepo,
		now:               now,
		tracer:            tracer,
	}
}
