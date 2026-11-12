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

type ZoneProblemNotifier struct {
	ch              cqs.CommandHandler
	title           string
	problemReason   string
	recoveryReasons map[string]bool

	mu           sync.Mutex
	failingZones map[string]bool
}

func (z *ZoneProblemNotifier) Listen(ctx context.Context, ev event.Event) error {
	var zone, reason string
	var detail *string
	switch e := ev.(type) {
	case *ml.WateringZoneSkippedEvent:
		zone, reason, detail = e.Zone, e.Reason, e.DecisionReason
	case *ml.WateringRequestedEvent:
		zone, reason, detail = e.Zone, e.Reason, e.DecisionReason
	default:
		return fmt.Errorf("invalid event type: %T", ev)
	}
	switch {
	case reason == z.problemReason:
		z.onFailure(ctx, zone, ptr.FromPointer(detail))
	case z.recoveryReasons[reason]:
		z.onRecovery(ctx, zone)
	}
	return nil
}

func (z *ZoneProblemNotifier) onFailure(ctx context.Context, zone, detail string) {
	z.mu.Lock()
	alreadyFailing := z.failingZones[zone]
	z.failingZones[zone] = true
	z.mu.Unlock()
	if alreadyFailing {
		return
	}

	message := fmt.Sprintf("WaterSystem\n%s ERROR\n", z.title)
	message += fmt.Sprintf("zone: %s\n", zone)
	message += fmt.Sprintf("error: %s\n", detail)
	message += "No more alerts will be sent for this zone until it recovers.\n"
	_, _ = z.ch.Handle(ctx, app.PublishMessageCommand{Message: message})
}

func (z *ZoneProblemNotifier) onRecovery(ctx context.Context, zone string) {
	z.mu.Lock()
	wasFailing := z.failingZones[zone]
	delete(z.failingZones, zone)
	z.mu.Unlock()
	if !wasFailing {
		return
	}

	message := fmt.Sprintf("WaterSystem\n%s RECOVERED\n", z.title)
	message += fmt.Sprintf("zone: %s\n", zone)
	_, _ = z.ch.Handle(ctx, app.PublishMessageCommand{Message: message})
}

func newZoneProblemNotifier(ch cqs.CommandHandler, title, problemReason string, recoveryReasons ...string) *ZoneProblemNotifier {
	recovery := make(map[string]bool, len(recoveryReasons))
	for _, r := range recoveryReasons {
		recovery[r] = true
	}
	return &ZoneProblemNotifier{
		ch:              ch,
		title:           title,
		problemReason:   problemReason,
		recoveryReasons: recovery,
		failingZones:    make(map[string]bool),
	}
}

func NewPublishMessageOnPredictionError(ch cqs.CommandHandler) *ZoneProblemNotifier {
	return newZoneProblemNotifier(ch, "ML Prediction", ml.PredictionErrorReason,
		ml.ModelNotEstimatedReason,
		ml.ModelPredictionReason,
	)
}

func NewPublishMessageOnSoilMeasureUnavailable(ch cqs.CommandHandler) *ZoneProblemNotifier {
	return newZoneProblemNotifier(ch, "Soil sensor", ml.SoilMeasureUnavailableReason, soilMeasureRecoveryReasons()...)
}

func NewPublishMessageOnSoilMeasureSuspicious(ch cqs.CommandHandler) *ZoneProblemNotifier {
	return newZoneProblemNotifier(ch, "Soil sensor out of range", ml.SoilMeasureSuspiciousReason,
		soilMeasureRecoveryReasons()...)
}

func soilMeasureRecoveryReasons() []string {
	return []string{
		ml.AboveMaxThresholdReason,
		ml.BelowMinThresholdReason,
		ml.ModelPredictionReason,
		ml.ModelNotEstimatedReason,
		ml.PredictionErrorReason,
		ml.PredictionPendingValidationReason,
		ml.IsNightRangeReason,
	}
}
