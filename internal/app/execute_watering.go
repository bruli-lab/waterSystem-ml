package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/domain/watering"
)

const ExecuteWateringCommandName = "execute_watering"

type ExecuteWateringCommand struct {
	Zone    string
	Seconds int
}

func (e ExecuteWateringCommand) Name() string {
	return ExecuteWateringCommandName
}

type ExecuteWatering struct {
	executor WateringExecutor
}

func (e ExecuteWatering) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(ExecuteWateringCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(ExecuteWateringCommandName, cmd.Name())
	}
	return nil, e.executor.Execute(ctx, watering.New(co.Zone, co.Seconds))
}

func NewExecuteWatering(executor WateringExecutor) *ExecuteWatering {
	return &ExecuteWatering{executor: executor}
}
