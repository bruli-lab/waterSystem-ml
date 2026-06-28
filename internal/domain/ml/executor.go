package ml

import "context"

type TrainExecutor interface {
	Run(ctx context.Context, zone string) error
}
