package ml

import (
	"time"

	"github.com/bruli-lab/go-core/event"
)

const (
	SystemDisabledReason              = "system_disabled"
	RainingReason                     = "raining"
	AboveMaxThresholdReason           = "above_max_threshold"
	BelowMinThresholdReason           = "below_min_threshold"
	ModelPredictionReason             = "model_prediction"
	ModelNotEstimatedReason           = "model_not_estimated"
	PredictionErrorReason             = "prediction_error"
	ZoneRecentlyExecutedByModelReason = "zone_recently_executed_by_model"
	IsNightRangeReason                = "is_night_range"
	PredictionPendingValidationReason = "prediction_pending_validation"
	SoilMeasureUnavailableReason      = "soil_measure_unavailable"
	SoilMeasureSuspiciousReason       = "soil_measure_suspicious"

	DefaultSecondsOnLowHumidity = 20

	SafetyWateringInterval = 6 * time.Hour
)

type CalculatedWatering struct {
	event.BasicAggregateRoot
	isRaining       bool
	systemActivated bool
	executions      Executions
	calculatedZones map[string]bool
	zonesHumidity   map[string]*ZoneHumidity
	ambient         AmbientTemperature
	timeFunc        func() time.Time
}

func (c *CalculatedWatering) Calculated() bool {
	for _, calc := range c.calculatedZones {
		if !calc {
			return false
		}
	}
	return true
}

func (c *CalculatedWatering) PendingPredictionZones() []string {
	zones := make([]string, 0)
	for zone, calculated := range c.calculatedZones {
		if !calculated {
			zones = append(zones, zone)
		}
	}
	return zones
}

func (c *CalculatedWatering) markAllCalculated() {
	for zone := range c.calculatedZones {
		c.calculatedZones[zone] = true
	}
}

func (c *CalculatedWatering) allowedFromSystem() bool {
	switch {
	case c.isRaining:
		c.Record(NewWateringSystemSkippedEvent(RainingReason))
		c.markAllCalculated()
		return false
	case !c.systemActivated:
		c.Record(NewWateringSystemSkippedEvent(SystemDisabledReason))
		c.markAllCalculated()
		return false
	default:
		return true
	}
}

func (c *CalculatedWatering) allowedFromZone(zone string, currentHumidity float64) bool {
	ex, ok := c.executions[zone]
	if ok && ex.IsRecentlyExecuted(c.timeFunc()) {
		c.Record(NewWateringZoneSkippedEvent(zone, ZoneRecentlyExecutedByModelReason, currentHumidity, nil, nil, nil))
		c.calculatedZones[zone] = true
		return false
	}
	return true
}

func (c *CalculatedWatering) isNightRange() bool {
	hour := c.timeFunc().Hour()
	return hour > 22 || hour <= 8
}

func (c *CalculatedWatering) FromPrediction(pred *Prediction, zh *ZoneHumidity) {
	if pred == nil {
		return
	}
	if c.calculatedZones[pred.Zone()] {
		return
	}
	if zh == nil {
		return
	}

	c.calculatedZones[pred.Zone()] = true

	if pred.Failed() {
		errorMessage := pred.ErrorMessage()
		c.Record(NewWateringZoneSkippedEvent(
			pred.Zone(),
			PredictionErrorReason,
			zh.CurrentHumidity(),
			nil,
			&errorMessage,
			nil,
		))
		return
	}

	predictionID := pred.ID()
	decisionReason := pred.DecisionReason()
	wateringProba := pred.WateringProba()

	if pred.shouldWater {
		c.Record(NewWateringRequestedEvent(
			pred.Zone(),
			ModelPredictionReason,
			pred.PredictedSeconds(),
			zh.CurrentHumidity(),
			zh.HumidityReference().TargetMoistureVoltage(),
			&predictionID,
			&decisionReason,
			&wateringProba,
		))
		return
	}

	c.Record(NewWateringZoneSkippedEvent(
		pred.Zone(),
		ModelNotEstimatedReason,
		zh.CurrentHumidity(),
		&predictionID,
		&decisionReason,
		&wateringProba,
	))
}

func (c *CalculatedWatering) safetyWatering(zh *ZoneHumidity, reason, cause string) {
	c.calculatedZones[zh.Zone()] = true
	if ex, ok := c.executions[zh.Zone()]; ok && c.timeFunc().Sub(ex.ExecutionTime()) < SafetyWateringInterval {
		c.skipSafetyWatering(zh, reason, cause+", watered less than 6h ago")
		return
	}
	if !c.ambient.IsHot() {
		c.skipSafetyWatering(zh, reason, cause+", not hot enough for a safety watering")
		return
	}
	decision := cause + ", safety watering"
	c.Record(NewWateringRequestedEvent(
		zh.Zone(),
		reason,
		DefaultSecondsOnLowHumidity,
		zh.CurrentHumidity(),
		zh.HumidityReference().TargetMoistureVoltage(),
		nil,
		&decision,
		nil,
	))
}

func (c *CalculatedWatering) skipSafetyWatering(zh *ZoneHumidity, reason, decision string) {
	c.Record(NewWateringZoneSkippedEvent(zh.Zone(), reason, zh.CurrentHumidity(), nil, &decision, nil))
}

func NewCalculatedWatering(
	isRaining bool,
	systemActivated bool,
	timeFunc func() time.Time,
	exec Executions,
	zonesHumidity []*ZoneHumidity,
	pendingPredictionLogsZones map[string]bool,
	ambient AmbientTemperature,
) *CalculatedWatering {
	calcZones := make(map[string]bool, len(zonesHumidity))
	zoneHumMap := make(map[string]*ZoneHumidity, len(zonesHumidity))
	for _, zh := range zonesHumidity {
		calcZones[zh.Zone()] = false
		zoneHumMap[zh.Zone()] = zh
	}

	cw := CalculatedWatering{
		isRaining:       isRaining,
		systemActivated: systemActivated,
		executions:      exec,
		timeFunc:        timeFunc,
		calculatedZones: calcZones,
		zonesHumidity:   zoneHumMap,
		ambient:         ambient,
	}

	if !cw.allowedFromSystem() {
		return &cw
	}

	for _, zh := range zonesHumidity {
		zone := zh.Zone()
		currentHumidity := zh.CurrentHumidity()
		if !zh.Measured() {
			cw.safetyWatering(zh, SoilMeasureUnavailableReason, "no recent soil moisture reading")
			continue
		}
		if zh.HumidityReference().IsSuspicious(currentHumidity) {
			cw.safetyWatering(zh, SoilMeasureSuspiciousReason, "soil moisture reading out of range")
			continue
		}
		if _, ok := pendingPredictionLogsZones[zone]; ok {
			cw.Record(NewWateringZoneSkippedEvent(
				zh.Zone(),
				PredictionPendingValidationReason,
				zh.CurrentHumidity(),
				nil,
				nil,
				nil,
			))
			cw.calculatedZones[zh.Zone()] = true
			continue
		}
		if !cw.allowedFromZone(zone, currentHumidity) {
			continue
		}

		switch {
		case zh.HumidityReference().IsHigh(currentHumidity):
			cw.Record(NewWateringZoneSkippedEvent(zone, AboveMaxThresholdReason, currentHumidity, nil, nil, nil))
			cw.calculatedZones[zone] = true
		case zh.HumidityReference().IsLow(currentHumidity):
			cw.Record(NewWateringRequestedEvent(
				zone,
				BelowMinThresholdReason,
				DefaultSecondsOnLowHumidity,
				currentHumidity,
				zh.HumidityReference().TargetMoistureVoltage(),
				nil,
				nil,
				nil,
			))
			cw.calculatedZones[zone] = true
		case cw.isNightRange():
			cw.Record(NewWateringZoneSkippedEvent(zone, IsNightRangeReason, currentHumidity, nil, nil, nil))
			cw.calculatedZones[zone] = true
		default:
		}
	}

	return &cw
}
