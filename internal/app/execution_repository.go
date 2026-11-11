package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type ExecutionRepository interface {
	GetLastExecution(ctx context.Context) (ml.Executions, error)
}
