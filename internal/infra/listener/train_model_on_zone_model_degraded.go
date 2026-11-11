package listener

import (
	"context"
	"fmt"

	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type TrainModelOnZoneModelDegraded struct {
	ch chan struct{ Zone string }
}

func (t TrainModelOnZoneModelDegraded) Listen(ctx context.Context, ev event.Event) error {
	zmd, ok := ev.(*ml.ZoneModelDegradedEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", ev)
	}
	select {
	case t.ch <- struct{ Zone string }{Zone: zmd.Zone}:
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func NewTrainModelOnZoneModelDegraded(ch chan struct{ Zone string }) *TrainModelOnZoneModelDegraded {
	return &TrainModelOnZoneModelDegraded{ch: ch}
}
