package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

const CalculateWateringEventName = "CalculateWatering"

type CalculateWateringCommand struct{}

func (c CalculateWateringCommand) Name() string {
	return CalculateWateringEventName
}

type CalculateWatering struct {
	svc *ml.Calculate
}

func (c CalculateWatering) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	_, ok := cmd.(CalculateWateringCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(CalculateWateringEventName, cmd.Name())
	}
	cw, err := c.svc.Do(ctx)
	if err != nil {
		return nil, err
	}
	return cw.Events(), nil
}

func NewCalculateWatering(svc *ml.Calculate) *CalculateWatering {
	return &CalculateWatering{svc: svc}
}
