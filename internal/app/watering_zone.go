package app

import (
	"context"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/domain/watering"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const WateringZoneCommandName = "watering_zone"

type WateringZoneCommand struct {
	Zone    string
	Seconds int
}

func (w WateringZoneCommand) Name() string {
	return WateringZoneCommandName
}

type WateringZone struct {
	svc    *watering.Execute
	tracer trace.Tracer
}

func (w WateringZone) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	ctx, span := w.tracer.Start(ctx, "app.WateringZone")
	defer span.End()
	co, ok := cmd.(WateringZoneCommand)
	if !ok {
		err := cqs.NewInvalidCommandError(WateringZoneCommandName, cmd.Name())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, w.svc.Execute(ctx, watering.New(co.Zone, co.Seconds))
}

func NewWateringZone(svc *watering.Execute, tracer trace.Tracer) *WateringZone {
	return &WateringZone{svc: svc, tracer: tracer}
}
