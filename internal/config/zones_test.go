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
	zones, err := config.LoadZones("../../manifests/zones.json")
	require.NoError(t, err)
	require.Equal(t, []string{"Bonsai big", "Bonsai small"}, zones.Names())
	require.Equal(t, map[string]string{
		"sensor.bonsai_big_bonsai_big_soil_moisture_voltage":     "Bonsai big",
		"sensor.bonsai_small_bonsai_small_soil_moisture_voltage": "Bonsai small",
	}, zones.MoistureVoltageMeasurements())
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
		"empty":            `{"zones": []}`,
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
		t.Run(name, func(t *testing.T) {
			_, err := config.LoadZones(writeZones(t, content))
			require.Error(t, err)
		})
	}
}

func TestLoadZones_MissingFile(t *testing.T) {
	_, err := config.LoadZones(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
}

func TestZone_CalibrationAt(t *testing.T) {
	zones, err := config.LoadZones(writeZones(t, `{"zones": [{"name": "a", "moisture_voltage_measurement": "m", "calibrations": [
		{"valid_from": "2026-09-01", "v40": 1.6, "v100": 1.4},
		{"valid_from": "2026-04-01", "v40": 1.5, "v100": 1.3}
	]}]}`))
	require.NoError(t, err)
	zone := zones[0]

	_, ok := zone.CalibrationAt(time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC))
	require.False(t, ok)

	c, ok := zone.CalibrationAt(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	require.True(t, ok)
	require.InDelta(t, 1.5, c.V40, 1e-9)

	c, ok = zone.CalibrationAt(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
	require.True(t, ok)
	require.InDelta(t, 1.6, c.V40, 1e-9)
}
