package listener

import (
	"context"
	"fmt"
	"sync"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/go-core/ptr"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type PublishMessageOnPredictionError struct {
	ch cqs.CommandHandler

	mu           sync.Mutex
	failingZones map[string]bool
}

func (p *PublishMessageOnPredictionError) Listen(ctx context.Context, ev event.Event) error {
	switch e := ev.(type) {
	case *ml.WateringZoneSkippedEvent:
		switch e.Reason {
		case ml.PredictionErrorReason:
			p.onFailure(ctx, e)
		case ml.ModelNotEstimatedReason:
			p.onRecovery(ctx, e.Zone)
		}
	case *ml.WateringRequestedEvent:
		if e.Reason == ml.ModelPredictionReason {
			p.onRecovery(ctx, e.Zone)
		}
	default:
		return fmt.Errorf("invalid event type: %T", ev)
	}
	return nil
}

func (p *PublishMessageOnPredictionError) onFailure(ctx context.Context, e *ml.WateringZoneSkippedEvent) {
	p.mu.Lock()
	alreadyFailing := p.failingZones[e.Zone]
	p.failingZones[e.Zone] = true
	p.mu.Unlock()
	if alreadyFailing {
		return
	}

	message := "WaterSystem\nML Prediction ERROR\n"
	message += fmt.Sprintf("zone: %s\n", e.Zone)
	message += fmt.Sprintf("error: %s\n", ptr.FromPointer(e.DecisionReason))
	message += fmt.Sprintf("soil moisture: %v\n", e.Moisture)
	message += "No more alerts will be sent for this zone until it recovers.\n"
	_, _ = p.ch.Handle(ctx, app.PublishMessageCommand{Message: message})
}

func (p *PublishMessageOnPredictionError) onRecovery(ctx context.Context, zone string) {
	p.mu.Lock()
	wasFailing := p.failingZones[zone]
	delete(p.failingZones, zone)
	p.mu.Unlock()
	if !wasFailing {
		return
	}

	message := "WaterSystem\nML Prediction RECOVERED\n"
	message += fmt.Sprintf("zone: %s\n", zone)
	_, _ = p.ch.Handle(ctx, app.PublishMessageCommand{Message: message})
}

func NewPublishMessageOnPredictionError(ch cqs.CommandHandler) *PublishMessageOnPredictionError {
	return &PublishMessageOnPredictionError{ch: ch, failingZones: make(map[string]bool)}
}
