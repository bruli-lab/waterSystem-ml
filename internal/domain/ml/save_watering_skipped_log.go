package ml

import "context"

type SaveWateringSkippedLog struct {
	repo WateringSkippedLogRepository
}

func (s SaveWateringSkippedLog) Save(ctx context.Context, ws *WateringSkippedLog) error {
	return s.repo.Save(ctx, ws)
}

func NewSaveWateringSkippedLog(repo WateringSkippedLogRepository) *SaveWateringSkippedLog {
	return &SaveWateringSkippedLog{repo: repo}
}
