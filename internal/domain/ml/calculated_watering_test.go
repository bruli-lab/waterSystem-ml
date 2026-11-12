package ml_test

import (
	"testing"
	"time"

	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/go-core/ptr"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	bonsaiBigZone   = "bonsai big"
	bonsaiSmallZone = "bonsai small"
	humidity100     = 1.2
	humidity40      = 1.8
	highHumidity    = 1.45
	lowHumidity     = 1.85
	mediumHumidity  = 1.7
	dryHumidity     = 2.2
	outOfRange      = 2.6
)

var dayTimeFunc = func() time.Time {
	return time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC)
}

var nightTimeFunc = func() time.Time {
	return time.Date(2021, 1, 1, 23, 0, 0, 0, time.UTC)
}

func oldExecution(zone string) ml.Execution {
	return ptr.FromPointer(ml.NewExecution(zone, dayTimeFunc().Add(-5*time.Hour)))
}

func staleExecution(zone string) ml.Execution {
	return ptr.FromPointer(ml.NewExecution(zone, dayTimeFunc().Add(-8*time.Hour)))
}

func recentExecution(zone string) ml.Execution {
	return ptr.FromPointer(ml.NewExecution(zone, dayTimeFunc().Add(-1*time.Hour)))
}

func executionsForAllZones() ml.Executions {
	return ml.Executions{
		bonsaiBigZone:   oldExecution(bonsaiBigZone),
		bonsaiSmallZone: oldExecution(bonsaiSmallZone),
	}
}

func humidityRef() *ml.HumidityReference {
	return ml.NewHumidityReference(humidity40, humidity100)
}

func zoneHumidity(zone string, humidity float64) *ml.ZoneHumidity {
	return ml.NewZoneHumidity(zone, humidity, humidityRef())
}

func reasonFromEvent(t *testing.T, ev event.Event) string {
	t.Helper()

	switch e := ev.(type) {
	case *ml.WateringRequestedEvent:
		return e.Reason
	case *ml.WateringZoneSkippedEvent:
		return e.Reason
	case *ml.WateringSystemSkippedEvent:
		return e.Reason
	default:
		t.Fatalf("unexpected event type: %T", ev)
		return ""
	}
}

func requireEvent(t *testing.T, ev, expectedEvent event.Event, expectedReason string) {
	t.Helper()

	require.IsType(t, expectedEvent, ev)
	require.Equal(t, expectedReason, reasonFromEvent(t, ev))
}

func TestNewCalculatedWatering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                       string
		isRaining                  bool
		systemActivated            bool
		timeFunc                   func() time.Time
		exec                       ml.Executions
		zonesHumidity              []*ml.ZoneHumidity
		pendingPredictionLogsZones map[string]bool
		ambient                    ml.AmbientTemperature
		expectedCalculated         bool
		expectedPendingZones       []string
		expectedEvents             []event.Event
		expectedReasons            []string
	}{
		{
			name:            "while it is raining, then it skips the whole system before checking zones",
			isRaining:       true,
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, lowHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringSystemSkippedEvent{},
			},
			expectedReasons: []string{
				ml.RainingReason,
			},
		},
		{
			name:            "with the system disabled, then it skips the whole system before checking zones",
			systemActivated: false,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, lowHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringSystemSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SystemDisabledReason,
			},
		},
		{
			name:            "at night with low and medium humidity, then it waters the low zone and skips the medium one",
			systemActivated: true,
			timeFunc:        nightTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, lowHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.BelowMinThresholdReason,
				ml.IsNightRangeReason,
			},
		},
		{
			name:            "at night with high humidity, then it skips the zone as above max threshold",
			systemActivated: true,
			timeFunc:        nightTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, highHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.AboveMaxThresholdReason,
			},
		},
		{
			name:            "with a pending validation in one zone, then it skips only that zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			pendingPredictionLogsZones: map[string]bool{
				bonsaiBigZone: true,
			},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, mediumHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated:   false,
			expectedPendingZones: []string{bonsaiSmallZone},
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.PredictionPendingValidationReason,
			},
		},
		{
			name:            "with high and low humidity, then it skips the high zone and waters the low one",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, highHumidity),
				zoneHumidity(bonsaiSmallZone, lowHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.AboveMaxThresholdReason,
				ml.BelowMinThresholdReason,
			},
		},
		{
			name:            "with a recent execution in one zone, then it skips that zone and waters the other low zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec: ml.Executions{
				bonsaiBigZone:   recentExecution(bonsaiBigZone),
				bonsaiSmallZone: oldExecution(bonsaiSmallZone),
			},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, lowHumidity),
				zoneHumidity(bonsaiSmallZone, lowHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.ZoneRecentlyExecutedByModelReason,
				ml.BelowMinThresholdReason,
			},
		},
		{
			name:            "with medium humidity, then the zone remains pending for prediction",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, mediumHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated:   false,
			expectedPendingZones: []string{bonsaiBigZone, bonsaiSmallZone},
		},
		{
			name:            "with a recent execution and medium humidity, then it skips the zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec: ml.Executions{
				bonsaiBigZone:   recentExecution(bonsaiBigZone),
				bonsaiSmallZone: oldExecution(bonsaiSmallZone),
			},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, mediumHumidity),
				zoneHumidity(bonsaiSmallZone, mediumHumidity),
			},
			expectedCalculated:   false,
			expectedPendingZones: []string{bonsaiSmallZone},
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.ZoneRecentlyExecutedByModelReason,
			},
		},
		{
			name:            "without recent soil measure and hot weather, then it requests a safety watering",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec: ml.Executions{
				bonsaiBigZone:   staleExecution(bonsaiBigZone),
				bonsaiSmallZone: oldExecution(bonsaiSmallZone),
			},
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
				zoneHumidity(bonsaiSmallZone, highHumidity),
			},
			ambient:            ml.NewAmbientTemperature(new(30.0), nil),
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
				ml.AboveMaxThresholdReason,
			},
		},
		{
			name:            "without recent soil measure and no execution history, then it requests a safety watering",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{},
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
			},
		},
		{
			name:            "without recent soil measure and no hot weather, then it skips the zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{bonsaiBigZone: staleExecution(bonsaiBigZone)},
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
			},
			ambient:            ml.NewAmbientTemperature(new(20.0), new(30.0)),
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
			},
		},
		{
			name:            "without recent soil measure and a watering less than 6h ago, then it skips the zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{bonsaiBigZone: oldExecution(bonsaiBigZone)},
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
			},
		},
		{
			name:            "without recent soil measure and a pending validation, then it still requests a safety watering",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{bonsaiBigZone: staleExecution(bonsaiBigZone)},
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
			},
			pendingPredictionLogsZones: map[string]bool{bonsaiBigZone: true},
			expectedCalculated:         true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
			},
		},
		{
			name:            "at night without recent soil measure, then it still requests a safety watering",
			systemActivated: true,
			timeFunc:        nightTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				ml.NewUnmeasuredZoneHumidity(bonsaiBigZone, humidityRef()),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureUnavailableReason,
			},
		},
		{
			name:            "with humidity below 15 percent, then it waters as low humidity",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, dryHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.BelowMinThresholdReason,
			},
		},
		{
			name:            "at night with humidity below 15 percent, then it waters as low humidity",
			systemActivated: true,
			timeFunc:        nightTimeFunc,
			exec:            executionsForAllZones(),
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, dryHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.BelowMinThresholdReason,
			},
		},
		{
			name:            "with an out of range reading and hot weather, then it requests a safety watering",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec: ml.Executions{
				bonsaiBigZone:   staleExecution(bonsaiBigZone),
				bonsaiSmallZone: staleExecution(bonsaiSmallZone),
			},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, outOfRange),
				zoneHumidity(bonsaiSmallZone, lowHumidity),
			},
			ambient:            ml.NewAmbientTemperature(new(30.0), nil),
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureSuspiciousReason,
				ml.BelowMinThresholdReason,
			},
		},
		{
			name:            "with an out of range reading and a watering less than 6h ago, then it skips the zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{bonsaiBigZone: oldExecution(bonsaiBigZone)},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, outOfRange),
			},
			ambient:            ml.NewAmbientTemperature(new(30.0), nil),
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureSuspiciousReason,
			},
		},
		{
			name:            "with an out of range reading and no hot weather, then it skips the zone",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec:            ml.Executions{bonsaiBigZone: staleExecution(bonsaiBigZone)},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, outOfRange),
			},
			ambient:            ml.NewAmbientTemperature(new(20.0), new(30.0)),
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.SoilMeasureSuspiciousReason,
			},
		},
		{
			name:            "without execution history, then the zone is evaluated as not recently watered",
			systemActivated: true,
			timeFunc:        dayTimeFunc,
			exec: ml.Executions{
				"wrong": oldExecution("wrong"),
			},
			zonesHumidity: []*ml.ZoneHumidity{
				zoneHumidity(bonsaiBigZone, lowHumidity),
			},
			expectedCalculated: true,
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.BelowMinThresholdReason,
			},
		},
	}

	for _, tt := range tests {
		t.Run(`Given a CalculatedWatering struct,
		when it is built `+tt.name, func(t *testing.T) {
			t.Parallel()

			got := ml.NewCalculatedWatering(
				tt.isRaining,
				tt.systemActivated,
				tt.timeFunc,
				tt.exec,
				tt.zonesHumidity,
				tt.pendingPredictionLogsZones,
				tt.ambient,
			)

			require.Equal(t, tt.expectedCalculated, got.Calculated())
			require.ElementsMatch(t, tt.expectedPendingZones, got.PendingPredictionZones())

			events := got.Events()
			require.Len(t, events, len(tt.expectedEvents))
			for i, ev := range events {
				requireEvent(t, ev, tt.expectedEvents[i], tt.expectedReasons[i])
			}
		})
	}
}

func TestCalculatedWatering_FromPrediction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		pred               *ml.Prediction
		zh                 *ml.ZoneHumidity
		expectedCalculated bool
		expectedEvents     []event.Event
		expectedReasons    []string
	}{
		{
			name: "with a prediction that says watering, then it requests watering",
			pred: ml.NewPrediction(uuid.New(), bonsaiBigZone, true, 10, "reason test", 0.5),
			zh:   zoneHumidity(bonsaiBigZone, mediumHumidity),
			expectedEvents: []event.Event{
				&ml.WateringRequestedEvent{},
			},
			expectedReasons: []string{
				ml.ModelPredictionReason,
			},
		},
		{
			name: "with a prediction that says no watering, then it skips the zone",
			pred: ml.NewPrediction(uuid.New(), bonsaiBigZone, false, 0, "reason test", 0.5),
			zh:   zoneHumidity(bonsaiBigZone, mediumHumidity),
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.ModelNotEstimatedReason,
			},
		},
		{
			name: "with a failed prediction, then it skips the zone with prediction error reason",
			pred: ml.NewFailedPrediction(uuid.New(), bonsaiBigZone, "Error loading model: boom"),
			zh:   zoneHumidity(bonsaiBigZone, mediumHumidity),
			expectedEvents: []event.Event{
				&ml.WateringZoneSkippedEvent{},
			},
			expectedReasons: []string{
				ml.PredictionErrorReason,
			},
		},
		{
			name: "without prediction, then it does nothing",
			pred: nil,
			zh:   zoneHumidity(bonsaiBigZone, mediumHumidity),
		},
		{
			name: "without zone humidity, then it does nothing",
			pred: ml.NewPrediction(uuid.New(), bonsaiBigZone, true, 10, "reason test", 0.5),
			zh:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(`Given a CalculatedWatering struct,
		when FromPrediction method is called `+tt.name, func(t *testing.T) {
			t.Parallel()

			cw := ml.NewCalculatedWatering(
				false,
				true,
				dayTimeFunc,
				executionsForAllZones(),
				[]*ml.ZoneHumidity{
					zoneHumidity(bonsaiBigZone, mediumHumidity),
					zoneHumidity(bonsaiSmallZone, mediumHumidity),
				},
				map[string]bool{},
				ml.NewAmbientTemperature(nil, nil),
			)
			require.False(t, cw.Calculated())

			cw.FromPrediction(tt.pred, tt.zh)

			events := cw.Events()
			require.Len(t, events, len(tt.expectedEvents))
			for i, ev := range events {
				requireEvent(t, ev, tt.expectedEvents[i], tt.expectedReasons[i])
			}
		})
	}
}
