package listener

import (
	"context"
	"fmt"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/app"
	"github.com/bruli/watersystem-ml/internal/domain/ml"
)

type TrainModelOnZoneModelDegraded struct {
	ch cqs.CommandHandler
}

func (t TrainModelOnZoneModelDegraded) Listen(ctx context.Context, ev event.Event) error {
	zmd, ok := ev.(*ml.ZoneModelDegradedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, _ = t.ch.Handle(ctx, app.TrainingZoneCommand{Zone: zmd.Zone})
	return nil
}

func NewTrainModelOnZoneModelDegraded(ch cqs.CommandHandler) *TrainModelOnZoneModelDegraded {
	return &TrainModelOnZoneModelDegraded{ch: ch}
}
