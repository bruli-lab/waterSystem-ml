package python

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const trainDryingFile = "python/train_drying.py"

type DryingTrainingExecutor struct {
	timeout    time.Duration
	tracer     trace.Tracer
	log        *slog.Logger
	pythonPath string
}

func (d DryingTrainingExecutor) Run(ctx context.Context, zone string, ifMissing bool) error {
	ctx, span := d.tracer.Start(ctx, "DryingTrainingExecutor.Run")
	defer span.End()
	args := []string{trainDryingFile, "--zone", zone}
	if ifMissing {
		args = append(args, "--if-missing")
	}
	if _, err := runCommand(ctx, d.timeout, d.pythonPath, args...); err != nil {
		err = fmt.Errorf("error running drying training for %q: %w", zone, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	d.log.InfoContext(ctx, "Drying training finished", slog.String("zone", zone))
	span.SetStatus(codes.Ok, "OK")
	return nil
}

func NewDryingTrainingExecutor(timeout time.Duration, pythonPath string, tracer trace.Tracer, log *slog.Logger) *DryingTrainingExecutor {
	return &DryingTrainingExecutor{timeout: timeout, tracer: tracer, log: log, pythonPath: pythonPath}
}
