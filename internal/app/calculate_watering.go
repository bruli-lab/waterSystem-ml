package app

import (
	"context"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
)

const CalculateWateringEventName = "CalculateWatering"

type CalculateWateringCommand struct{}

func (c CalculateWateringCommand) Name() string {
	return CalculateWateringEventName
}

type CalculateWatering struct {
	predictionRepo    PredictionRepository
	soilMeasureRepo   SoilMeasureRepository
	humidityRefRepo   HumidityReferenceRepository
	executionRepo     ExecutionRepository
	predictionLogRepo PredictionLogRepository
	statusRepo        StatusRepository
	now               func() time.Time
}

func (c *CalculateWatering) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	if _, ok := cmd.(CalculateWateringCommand); !ok {
		return nil, cqs.NewInvalidCommandError(CalculateWateringEventName, cmd.Name())
	}
	cw, err := c.calculate(ctx)
	if err != nil {
		return nil, err
	}
	return cw.Events(), nil
}

func (c *CalculateWatering) calculate(ctx context.Context) (*ml.CalculatedWatering, error) {
	st, err := c.statusRepo.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	exec, err := c.executionRepo.GetLastExecution(ctx)
	if err != nil {
		return nil, err
	}
	measures, err := c.soilMeasureRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	pendingPredictionLogs, err := c.predictionLogRepo.GetPendingValidationZones(ctx)
	if err != nil {
		return nil, err
	}
	zoneHum, zoneHumMap, err := c.zonesHumidity(ctx, measures)
	if err != nil {
		return nil, err
	}

	cw, err := ml.NewCalculatedWatering(st.Raining(), st.Active(), c.now, exec, zoneHum, pendingPredictionLogs)
	if err != nil {
		return nil, err
	}
	if cw.Calculated() {
		return cw, nil
	}

	predictions, err := c.predictionRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	for i := range predictions {
		zh, ok := zoneHumMap[predictions[i].Zone()]
		if !ok {
			continue
		}
		cw.FromPrediction(&predictions[i], zh)
	}
	return cw, nil
}

func (c *CalculateWatering) zonesHumidity(
	ctx context.Context,
	measures []ml.SoilMeasure,
) ([]*ml.ZoneHumidity, map[string]*ml.ZoneHumidity, error) {
	zoneHumMap := make(map[string]*ml.ZoneHumidity, len(measures))
	zoneHum := make([]*ml.ZoneHumidity, 0, len(measures))
	for _, measure := range measures {
		ref, err := c.humidityRefRepo.GetByZone(ctx, measure.Zone())
		if err != nil {
			return nil, nil, err
		}
		hum := ml.NewZoneHumidity(measure.Zone(), measure.Humidity(), ref)
		zoneHum = append(zoneHum, hum)
		zoneHumMap[measure.Zone()] = hum
	}
	return zoneHum, zoneHumMap, nil
}

func NewCalculateWatering(
	predictionRepo PredictionRepository,
	soilMeasureRepo SoilMeasureRepository,
	humidityRefRepo HumidityReferenceRepository,
	executionRepo ExecutionRepository,
	predictionLogRepo PredictionLogRepository,
	statusRepo StatusRepository,
	now func() time.Time,
) *CalculateWatering {
	return &CalculateWatering{
		predictionRepo:    predictionRepo,
		soilMeasureRepo:   soilMeasureRepo,
		humidityRefRepo:   humidityRefRepo,
		executionRepo:     executionRepo,
		predictionLogRepo: predictionLogRepo,
		statusRepo:        statusRepo,
		now:               now,
	}
}
