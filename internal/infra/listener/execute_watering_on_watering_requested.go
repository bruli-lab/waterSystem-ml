package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/go-core/ptr"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ExecuteWateringOnWateringRequested struct {
	ch cqs.CommandHandler
}

func (e ExecuteWateringOnWateringRequested) Listen(ctx context.Context, ev event.Event) error {
	swr, ok := ev.(*ml.WateringRequestedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, err := e.ch.Handle(ctx, app.ExecuteWateringCommand{
		Zone:    swr.Zone,
		Seconds: int(swr.Seconds),
	})
	e.savePredictionLog(ctx, swr, err == nil)
	if err != nil {
		_, _ = e.ch.Handle(ctx, app.PublishMessageCommand{
			Message: fmt.Sprintf("error executing watering on zone %s: %s", swr.Zone, err.Error()),
		})
	}
	return nil
}

func (e ExecuteWateringOnWateringRequested) savePredictionLog(ctx context.Context, swr *ml.WateringRequestedEvent, executed bool) {
	if swr.PredictionID == nil {
		return
	}
	_, _ = e.ch.Handle(ctx, app.SavePredictionLogCommand{
		ID:               ptr.FromPointer(swr.PredictionID),
		Zone:             swr.Zone,
		ShouldWater:      true,
		PredictedSeconds: swr.Seconds,
		DecisionReason:   ptr.FromPointer(swr.DecisionReason),
		MoistureBefore:   swr.MoistureBefore,
		WateringExecuted: executed,
		TargetMoisture:   swr.TargetMoisture,
	})
}

func NewExecuteWateringOnWateringRequested(ch cqs.CommandHandler) *ExecuteWateringOnWateringRequested {
	return &ExecuteWateringOnWateringRequested{ch: ch}
}
