package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli/go-core/cqs"
	"github.com/bruli/go-core/event"
)

type CheckModelOnPredictionValidationFailed struct {
	ch cqs.CommandHandler
}

func (c CheckModelOnPredictionValidationFailed) Listen(ctx context.Context, ev event.Event) error {
	pvf, ok := ev.(*ml.PredictionValidationFailedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	_, _ = c.ch.Handle(ctx, app.CheckFailedModelCommand{Zone: pvf.Zone})
	return nil
}

func NewCheckModelOnPredictionValidationFailed(ch cqs.CommandHandler) *CheckModelOnPredictionValidationFailed {
	return &CheckModelOnPredictionValidationFailed{ch: ch}
}
