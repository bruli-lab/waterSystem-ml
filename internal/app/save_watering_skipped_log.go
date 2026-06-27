package app

import (
	"context"
	"time"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const SaveWateringSkippedLogCommandName = "save_watering_skipped_log"

type SaveWateringSkippedLogCommand struct {
	Zone           *string
	Reason         string
	Moisture       float64
	PredictionID   *uuid.UUID
	DecisionReason *string
	WateringProba  *float64
	ExecutedAt     time.Time
}

func (w *SaveWateringSkippedLogCommand) Name() string {
	return SaveWateringSkippedLogCommandName
}

type SaveWateringSkippedLog struct {
	svc    *ml.SaveWateringSkippedLog
	tracer trace.Tracer
}

func (w SaveWateringSkippedLog) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	ctx, span := w.tracer.Start(ctx, "SaveWateringSkippedLog")
	defer span.End()
	co, ok := cmd.(*SaveWateringSkippedLogCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(SaveWateringSkippedLogCommandName, cmd.Name())
	}
	ws := ml.NewWateringSkippedLog(co.Zone, co.Reason, co.Moisture, co.PredictionID, co.DecisionReason, co.WateringProba, co.ExecutedAt)
	if err := w.svc.Save(ctx, ws); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, nil
}

func NewSaveWateringSkippedLog(svc *ml.SaveWateringSkippedLog, tracer trace.Tracer) *SaveWateringSkippedLog {
	return &SaveWateringSkippedLog{svc: svc, tracer: tracer}
}
