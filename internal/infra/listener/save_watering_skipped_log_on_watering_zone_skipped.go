package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type SaveWateringSkippedLogOnWateringZoneSkipped struct {
	ch cqs.CommandHandler
}

func (s SaveWateringSkippedLogOnWateringZoneSkipped) Listen(ctx context.Context, ev event.Event) error {
	wzk, ok := ev.(*ml.WateringZoneSkippedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, _ = s.ch.Handle(ctx, &app.SaveWateringSkippedLogCommand{
		Zone:           new(wzk.Zone),
		Reason:         wzk.Reason,
		Moisture:       new(wzk.Moisture),
		PredictionID:   wzk.PredictionID,
		DecisionReason: wzk.DecisionReason,
		WateringProba:  wzk.WateringProba,
	})
	return nil
}

func NewSaveWateringSkippedLogOnWateringZoneSkipped(ch cqs.CommandHandler) *SaveWateringSkippedLogOnWateringZoneSkipped {
	return &SaveWateringSkippedLogOnWateringZoneSkipped{ch: ch}
}
