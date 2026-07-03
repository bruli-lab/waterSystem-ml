package ml

import "context"

type SaveModelTrainingState struct {
	repo ModelTrainingStateRepository
}

func (s SaveModelTrainingState) Save(ctx context.Context, state *ModelTrainingState) error {
	return s.repo.Save(ctx, state)
}

func NewSaveModelTrainingState(repo ModelTrainingStateRepository) *SaveModelTrainingState {
	return &SaveModelTrainingState{repo: repo}
}
