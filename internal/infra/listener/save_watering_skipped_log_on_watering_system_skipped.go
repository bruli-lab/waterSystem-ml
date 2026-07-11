package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type SaveWateringSkippedLogOnWateringSystemSkipped struct {
	ch cqs.CommandHandler
}

func (s SaveWateringSkippedLogOnWateringSystemSkipped) Listen(ctx context.Context, ev event.Event) error {
	wsk, ok := ev.(*ml.WateringSystemSkippedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, _ = s.ch.Handle(ctx, &app.SaveWateringSkippedLogCommand{
		Reason: wsk.Reason,
	})
	return nil
}

func NewSaveWateringSkippedLogOnWateringSystemSkipped(ch cqs.CommandHandler) *SaveWateringSkippedLogOnWateringSystemSkipped {
	return &SaveWateringSkippedLogOnWateringSystemSkipped{ch: ch}
}
