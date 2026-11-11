package ml

import (
	"errors"
	"time"
)

var ErrInvalidShadowHorizon = errors.New("shadow horizon must be at least one hour")

type HumidityEstimate struct {
	pessimistic float64
	median      float64
}

func (h HumidityEstimate) Pessimistic() float64 {
	return h.pessimistic
}

func (h HumidityEstimate) Median() float64 {
	return h.median
}

func NewHumidityEstimate(pessimistic, median float64) HumidityEstimate {
	return HumidityEstimate{pessimistic: pessimistic, median: median}
}

type ShadowHorizon struct {
	hours int
	night bool
}

func (s ShadowHorizon) Hours() int {
	return s.hours
}

func (s ShadowHorizon) Night() bool {
	return s.night
}

func NewShadowHorizon(hours int, night bool) (ShadowHorizon, error) {
	if hours < 1 {
		return ShadowHorizon{}, ErrInvalidShadowHorizon
	}
	return ShadowHorizon{hours: hours, night: night}, nil
}

type ShadowTrajectory struct {
	humidity           float64
	startHumidity      float64
	horizon            ShadowHorizon
	minimum            HumidityEstimate
	at1h               HumidityEstimate
	at3h               HumidityEstimate
	rainProbabilityMax *float64
	modelTrainedAt     string
}

func (s *ShadowTrajectory) Humidity() float64 {
	return s.humidity
}

func (s *ShadowTrajectory) StartHumidity() float64 {
	return s.startHumidity
}

func (s *ShadowTrajectory) Horizon() ShadowHorizon {
	return s.horizon
}

func (s *ShadowTrajectory) Minimum() HumidityEstimate {
	return s.minimum
}

func (s *ShadowTrajectory) At1h() HumidityEstimate {
	return s.at1h
}

func (s *ShadowTrajectory) At3h() HumidityEstimate {
	return s.at3h
}

func (s *ShadowTrajectory) RainProbabilityMax() *float64 {
	return s.rainProbabilityMax
}

func (s *ShadowTrajectory) ModelTrainedAt() string {
	return s.modelTrainedAt
}

func NewShadowTrajectory(
	humidity, startHumidity float64,
	horizon ShadowHorizon,
	minimum, at1h, at3h HumidityEstimate,
	rainProbabilityMax *float64,
	modelTrainedAt string,
) *ShadowTrajectory {
	return &ShadowTrajectory{
		humidity:           humidity,
		startHumidity:      startHumidity,
		horizon:            horizon,
		minimum:            minimum,
		at1h:               at1h,
		at3h:               at3h,
		rainProbabilityMax: rainProbabilityMax,
		modelTrainedAt:     modelTrainedAt,
	}
}

type ShadowPrediction struct {
	zone        string
	predictedAt time.Time
	trajectory  *ShadowTrajectory
	wouldWater  bool
	err         string
	humidity    *float64
}

func (s *ShadowPrediction) Zone() string {
	return s.zone
}

func (s *ShadowPrediction) PredictedAt() time.Time {
	return s.predictedAt
}

func (s *ShadowPrediction) Trajectory() *ShadowTrajectory {
	return s.trajectory
}

func (s *ShadowPrediction) WouldWater() bool {
	return s.wouldWater
}

func (s *ShadowPrediction) Error() string {
	return s.err
}

func (s *ShadowPrediction) Failed() bool {
	return s.err != ""
}

func (s *ShadowPrediction) Humidity() *float64 {
	if s.trajectory != nil {
		humidity := s.trajectory.Humidity()
		return &humidity
	}
	return s.humidity
}

func NewShadowPrediction(zone string, predictedAt time.Time, trajectory *ShadowTrajectory, wouldWater bool) *ShadowPrediction {
	return &ShadowPrediction{zone: zone, predictedAt: predictedAt, trajectory: trajectory, wouldWater: wouldWater}
}

func NewFailedShadowPrediction(zone string, predictedAt time.Time, err string, humidity *float64) *ShadowPrediction {
	return &ShadowPrediction{zone: zone, predictedAt: predictedAt, err: err, humidity: humidity}
}
