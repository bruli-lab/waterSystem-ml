package python

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const shadowFile = "python/shadow.py"

type ShadowResult struct {
	Zone               string   `json:"zone"`
	Humidity           *float64 `json:"humidity"`
	StartHumidity      float64  `json:"start_humidity"`
	HorizonHours       int      `json:"horizon_hours"`
	Night              bool     `json:"night"`
	MinPessimistic     float64  `json:"min_pessimistic"`
	MinMedian          float64  `json:"min_median"`
	Pessimistic1h      float64  `json:"pessimistic_1h"`
	Pessimistic3h      float64  `json:"pessimistic_3h"`
	Median1h           float64  `json:"median_1h"`
	Median3h           float64  `json:"median_3h"`
	RainProbabilityMax *float64 `json:"rain_probability_max"`
	WouldWater         bool     `json:"would_water"`
	ModelTrainedAt     string   `json:"model_trained_at"`
	Error              string   `json:"error"`
}

func parseShadowResults(output string) ([]ShadowResult, error) {
	var results []ShadowResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		return nil, fmt.Errorf("error parsing shadow JSON: %w; output=%q", err, output)
	}
	return results, nil
}

func buildShadowDomain(results []ShadowResult, now time.Time) []ml.ShadowPrediction {
	predictions := make([]ml.ShadowPrediction, len(results))
	for i := range results {
		predictions[i] = *buildShadowPrediction(&results[i], now)
	}
	return predictions
}

func buildShadowPrediction(r *ShadowResult, now time.Time) *ml.ShadowPrediction {
	if r.Error != "" {
		return ml.NewFailedShadowPrediction(r.Zone, now, r.Error, r.Humidity)
	}
	if r.Humidity == nil {
		return ml.NewFailedShadowPrediction(r.Zone, now, "missing humidity", nil)
	}
	horizon, err := ml.NewShadowHorizon(r.HorizonHours, r.Night)
	if err != nil {
		return ml.NewFailedShadowPrediction(r.Zone, now, err.Error(), r.Humidity)
	}
	trajectory := ml.NewShadowTrajectory(
		*r.Humidity,
		r.StartHumidity,
		horizon,
		ml.NewHumidityEstimate(r.MinPessimistic, r.MinMedian),
		ml.NewHumidityEstimate(r.Pessimistic1h, r.Median1h),
		ml.NewHumidityEstimate(r.Pessimistic3h, r.Median3h),
		r.RainProbabilityMax,
		r.ModelTrainedAt,
	)
	return ml.NewShadowPrediction(r.Zone, now, trajectory, r.WouldWater)
}

type ShadowPredictor struct {
	tracer     trace.Tracer
	pythonPath string
	timeout    time.Duration
}

func (s ShadowPredictor) Predict(ctx context.Context, now time.Time) ([]ml.ShadowPrediction, error) {
	ctx, span := s.tracer.Start(ctx, "ShadowPredictor.Predict")
	defer span.End()
	output, err := runCommand(ctx, s.timeout, s.pythonPath, shadowFile)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error running shadow prediction: %w", err)
	}
	results, err := parseShadowResults(output)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "OK")
	return buildShadowDomain(results, now), nil
}

func NewShadowPredictor(tracer trace.Tracer, pythonPath string, timeout time.Duration) *ShadowPredictor {
	return &ShadowPredictor{tracer: tracer, pythonPath: pythonPath, timeout: timeout}
}
