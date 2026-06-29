package ml

import (
	"context"
	"time"
)

type SaveModelTrainingLog struct {
	repo ModelTrainingLogRepository
}

func (s SaveModelTrainingLog) Save(ctx context.Context, zone string, at time.Time) error {
	return s.repo.Save(ctx, zone, at)
}

func NewSaveModelTrainingLog(repo ModelTrainingLogRepository) *SaveModelTrainingLog {
	return &SaveModelTrainingLog{repo: repo}
}
