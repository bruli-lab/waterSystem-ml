package ml

import "github.com/google/uuid"

type Prediction struct {
	id               uuid.UUID
	zone             string
	shouldWater      bool
	predictedSeconds float64
	decisionReason   string
	wateringProba    float64
}

func (p Prediction) WateringProba() float64 {
	return p.wateringProba
}

func (p Prediction) ID() uuid.UUID {
	return p.id
}

func (p Prediction) Zone() string {
	return p.zone
}

func (p Prediction) ShouldWater() bool {
	return p.shouldWater
}

func (p Prediction) PredictedSeconds() float64 {
	return p.predictedSeconds
}

func (p Prediction) DecisionReason() string {
	return p.decisionReason
}

func NewPrediction(
	id uuid.UUID,
	zone string,
	shouldWater bool,
	predictedSeconds float64,
	decisionReason string,
	wateringProba float64,
) *Prediction {
	return &Prediction{
		id:               id,
		zone:             zone,
		shouldWater:      shouldWater,
		predictedSeconds: predictedSeconds,
		decisionReason:   decisionReason,
		wateringProba:    wateringProba,
	}
}
