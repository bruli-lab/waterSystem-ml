package ml

const (
	HotForecastTemperature = 25.0
	HotSensorTemperature   = 31.0
)

type AmbientTemperature struct {
	forecast *float64
	sensor   *float64
}

func (a AmbientTemperature) IsHot() bool {
	switch {
	case a.forecast != nil:
		return *a.forecast >= HotForecastTemperature
	case a.sensor != nil:
		return *a.sensor >= HotSensorTemperature
	default:
		return true
	}
}

func NewAmbientTemperature(forecast, sensor *float64) AmbientTemperature {
	return AmbientTemperature{forecast: forecast, sensor: sensor}
}
