package ml

import "time"

type ModelTrainingState struct {
	zone       string
	trainingAt time.Time
}

func (m ModelTrainingState) Zone() string {
	return m.zone
}

func (m ModelTrainingState) TrainingAt() time.Time {
	return m.trainingAt
}

func (m ModelTrainingState) IsRecentlyTraining() bool {
	return time.Since(m.trainingAt) <= 48*time.Hour
}

func NewModelTrainingState(zone string, trainingAt time.Time) *ModelTrainingState {
	return &ModelTrainingState{zone: zone, trainingAt: trainingAt}
}
