package listener

import (
	"context"
	"fmt"

	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
	"github.com/bruli/watersystem-ml/internal/app"
	"github.com/bruli/watersystem-ml/internal/domain/ml"
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
