package app

import "context"

type DryingTrainExecutor interface {
	Run(ctx context.Context, zone string, ifMissing bool) error
}
