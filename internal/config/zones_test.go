package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/config"
	"github.com/stretchr/testify/require"
)

func writeZones(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "zones.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadZones_ManifestFile(t *testing.T) {
	t.Run(`Given the zones manifest file,
		when LoadZones is called, then it returns both bonsai zones with their measurements`, func(t *testing.T) {
		zones, err := config.LoadZones("../../manifests/zones.json")
		require.NoError(t, err)
		require.Equal(t, []string{"Bonsai big", "Bonsai small"}, zones.Names())
		require.Equal(t, map[string]string{
			"sensor.bonsai_big_bonsai_big_soil_moisture_voltage":     "Bonsai big",
			"sensor.bonsai_small_bonsai_small_soil_moisture_voltage": "Bonsai small",
		}, zones.MoistureVoltageMeasurements())
	})
}

func zoneJSON(name, measurement, validFrom string, v40 float64) string {
	return fmt.Sprintf(
		`{"name": %q, "moisture_voltage_measurement": %q, "calibrations": [{"valid_from": %q, "v40": %v, "v100": 1.3}]}`,
		name, measurement, validFrom, v40,
	)
}

func TestLoadZones_Invalid(t *testing.T) {
	valid := zoneJSON("a", "m", "2026-01-01", 1.5)
	tests := map[string]string{
		"no zones":         `{"zones": []}`,
		"no name":          `{"zones": [` + zoneJSON("", "m", "2026-01-01", 1.5) + `]}`,
		"no measurement":   `{"zones": [` + zoneJSON("a", "", "2026-01-01", 1.5) + `]}`,
		"no calibrations":  `{"zones": [{"name": "a", "moisture_voltage_measurement": "m", "calibrations": []}]}`,
		"no valid_from":    `{"zones": [{"name": "a", "moisture_voltage_measurement": "m", "calibrations": [{"v40": 1.5, "v100": 1.3}]}]}`,
		"bad date":         `{"zones": [` + zoneJSON("a", "m", "01/01/2026", 1.5) + `]}`,
		"v40 below v100":   `{"zones": [` + zoneJSON("a", "m", "2026-01-01", 1.2) + `]}`,
		"duplicated zones": `{"zones": [` + valid + `, ` + valid + `]}`,
		"malformed json":   `{"zones": [`,
	}
	for name, content := range tests {
		t.Run(`Given an invalid zones file,
		when LoadZones is called with `+name+`, then it returns an error`, func(t *testing.T) {
			_, err := config.LoadZones(writeZones(t, content))
			require.Error(t, err)
		})
	}
}

func TestLoadZones_MissingFile(t *testing.T) {
	t.Run(`Given a path without file,
		when LoadZones is called, then it returns an error`, func(t *testing.T) {
		_, err := config.LoadZones(filepath.Join(t.TempDir(), "missing.json"))
		require.Error(t, err)
	})
}

func TestZone_CalibrationAt(t *testing.T) {
	zones, err := config.LoadZones(writeZones(t, `{"zones": [{"name": "a", "moisture_voltage_measurement": "m", "calibrations": [
		{"valid_from": "2026-09-01", "v40": 1.6, "v100": 1.4},
		{"valid_from": "2026-04-01", "v40": 1.5, "v100": 1.3}
	]}]}`))
	require.NoError(t, err)
	zone := zones[0]

	tests := []struct {
		name  string
		at    time.Time
		found bool
		v40   float64
	}{
		{
			name: "with a date before the first calibration, then it returns no calibration",
			at:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "with a date between calibrations, then it returns the older one",
			at:    time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			found: true,
			v40:   1.5,
		},
		{
			name:  "with a date after the last calibration, then it returns the newest one",
			at:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			found: true,
			v40:   1.6,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a Zone with two calibrations,
		when CalibrationAt method is called `+tt.name, func(t *testing.T) {
			c, ok := zone.CalibrationAt(tt.at)
			require.Equal(t, tt.found, ok)
			if tt.found {
				require.InDelta(t, tt.v40, c.V40, 1e-9)
			}
		})
	}
}
