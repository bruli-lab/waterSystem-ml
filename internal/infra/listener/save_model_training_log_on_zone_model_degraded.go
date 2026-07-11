package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
)

type SaveModelTrainingLogOnZoneModelDegraded struct {
	ch cqs.CommandHandler
}

func (s SaveModelTrainingLogOnZoneModelDegraded) Listen(ctx context.Context, ev event.Event) error {
	zmd, ok := ev.(*ml.ZoneModelDegradedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, _ = s.ch.Handle(ctx, app.SaveModelTrainingLogCommand{
		Zone:       zmd.Zone,
		ExecutedAt: zmd.EventAt(),
	})
	return nil
}

func NewSaveModelTrainingLogOnZoneModelDegraded(ch cqs.CommandHandler) *SaveModelTrainingLogOnZoneModelDegraded {
	return &SaveModelTrainingLogOnZoneModelDegraded{ch: ch}
}
