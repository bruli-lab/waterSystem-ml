package ml

import (
	"time"

	"github.com/google/uuid"
)

type WateringSkippedLog struct {
	zone           *string
	reason         string
	moisture       *float64
	predictionID   *uuid.UUID
	decisionReason *string
	wateringProba  *float64
	executedAt     time.Time
}

func (w *WateringSkippedLog) ExecutedAt() time.Time {
	return w.executedAt
}

func (w *WateringSkippedLog) Zone() *string {
	return w.zone
}

func (w *WateringSkippedLog) Reason() string {
	return w.reason
}

func (w *WateringSkippedLog) Moisture() *float64 {
	return w.moisture
}

func (w *WateringSkippedLog) PredictionID() *uuid.UUID {
	return w.predictionID
}

func (w *WateringSkippedLog) DecisionReason() *string {
	return w.decisionReason
}

func (w *WateringSkippedLog) WateringProba() *float64 {
	return w.wateringProba
}

func NewWateringSkippedLog(
	zone *string,
	reason string,
	moisture *float64,
	predictionID *uuid.UUID,
	decisionReason *string,
	wateringProba *float64,
	executedAt time.Time,
) *WateringSkippedLog {
	return &WateringSkippedLog{
		zone:           zone,
		reason:         reason,
		moisture:       moisture,
		predictionID:   predictionID,
		decisionReason: decisionReason,
		wateringProba:  wateringProba,
		executedAt:     executedAt,
	}
}
