package app

import (
	"context"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
)

const RunShadowPredictionCommandName = "run_shadow_prediction"

type RunShadowPredictionCommand struct{}

func (r RunShadowPredictionCommand) Name() string {
	return RunShadowPredictionCommandName
}

type RunShadowPrediction struct {
	predictor ShadowPredictor
	repo      ShadowPredictionRepository
	now       func() time.Time
}

func (r RunShadowPrediction) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	if _, ok := cmd.(RunShadowPredictionCommand); !ok {
		return nil, cqs.NewInvalidCommandError(RunShadowPredictionCommandName, cmd.Name())
	}
	predictions, err := r.predictor.Predict(ctx, r.now())
	if err != nil {
		return nil, err
	}
	if len(predictions) == 0 {
		return nil, nil
	}
	return nil, r.repo.Save(ctx, predictions)
}

func NewRunShadowPrediction(predictor ShadowPredictor, repo ShadowPredictionRepository, now func() time.Time) *RunShadowPrediction {
	return &RunShadowPrediction{predictor: predictor, repo: repo, now: now}
}
