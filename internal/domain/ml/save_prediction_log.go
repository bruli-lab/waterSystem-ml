package ml

import "context"

type SavePredictionLog struct {
	repo PredictionLogRepository
}

func (spl SavePredictionLog) Save(ctx context.Context, pl *PredictionLog) error {
	return spl.repo.Save(ctx, pl)
}

func NewSavePredictionLog(repo PredictionLogRepository) *SavePredictionLog {
	return &SavePredictionLog{repo: repo}
}
