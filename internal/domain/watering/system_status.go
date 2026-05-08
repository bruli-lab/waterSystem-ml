package watering

import "context"

type Status struct {
	active bool
}

func (s Status) Active() bool {
	return s.active
}

func NewStatus(active bool) *Status {
	return &Status{active: active}
}

type StatusRepository interface {
	GetStatus(ctx context.Context) (*Status, error)
}
type SystemStatus struct {
	repository StatusRepository
}

func (s SystemStatus) GetStatus(ctx context.Context) (*Status, error) {
	return s.repository.GetStatus(ctx)
}

func NewSystemStatus(repository StatusRepository) *SystemStatus {
	return &SystemStatus{repository: repository}
}
