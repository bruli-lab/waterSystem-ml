package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", raw, err)
	}
	d.Time = t
	return nil
}

type Calibration struct {
	ValidFrom Date    `json:"valid_from"`
	V40       float64 `json:"v40"`
	V100      float64 `json:"v100"`
}

type Zone struct {
	Name                       string        `json:"name"`
	MoistureVoltageMeasurement string        `json:"moisture_voltage_measurement"`
	SoilTemperatureMeasurement string        `json:"soil_temperature_measurement"`
	Calibrations               []Calibration `json:"calibrations"`
}

func (z Zone) CalibrationAt(at time.Time) (Calibration, bool) {
	var current Calibration
	found := false
	for _, c := range z.Calibrations {
		if c.ValidFrom.After(at) {
			continue
		}
		if !found || c.ValidFrom.After(current.ValidFrom.Time) {
			current = c
			found = true
		}
	}
	return current, found
}

type Zones []Zone

func (z Zones) Names() []string {
	names := make([]string, 0, len(z))
	for _, zone := range z {
		names = append(names, zone.Name)
	}
	return names
}

func (z Zones) MoistureVoltageMeasurements() map[string]string {
	measurements := make(map[string]string, len(z))
	for _, zone := range z {
		measurements[zone.MoistureVoltageMeasurement] = zone.Name
	}
	return measurements
}

type zonesFile struct {
	Zones Zones `json:"zones"`
}

func LoadZones(path string) (Zones, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("error reading zones file: %w", err)
	}
	var file zonesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("error parsing zones file: %w", err)
	}
	if err := file.Zones.validate(); err != nil {
		return nil, fmt.Errorf("invalid zones file: %w", err)
	}
	return file.Zones, nil
}

func (z Zones) validate() error {
	if len(z) == 0 {
		return errors.New("no zones defined")
	}
	names := make([]string, 0, len(z))
	for _, zone := range z {
		if err := zone.validate(); err != nil {
			return err
		}
		if slices.Contains(names, zone.Name) {
			return fmt.Errorf("duplicated zone %q", zone.Name)
		}
		names = append(names, zone.Name)
	}
	return nil
}

func (z Zone) validate() error {
	if strings.TrimSpace(z.Name) == "" {
		return errors.New("zone without name")
	}
	if z.MoistureVoltageMeasurement == "" {
		return fmt.Errorf("zone %q without moisture_voltage_measurement", z.Name)
	}
	if len(z.Calibrations) == 0 {
		return fmt.Errorf("zone %q without calibrations", z.Name)
	}
	for _, c := range z.Calibrations {
		if c.ValidFrom.IsZero() {
			return fmt.Errorf("zone %q has a calibration without valid_from", z.Name)
		}
		if c.V100 <= 0 || c.V40 <= c.V100 {
			return fmt.Errorf("zone %q has an invalid calibration: v40 must be greater than v100", z.Name)
		}
	}
	return nil
}
