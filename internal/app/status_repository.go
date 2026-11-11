package app

import (
	"context"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

type StatusRepository interface {
	GetStatus(ctx context.Context) (*ml.Status, error)
}
