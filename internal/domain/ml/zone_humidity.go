package ml

type ZoneHumidity struct {
	zone              string
	currentHumidity   float64
	measured          bool
	humidityReference *HumidityReference
}

func (z ZoneHumidity) HumidityReference() *HumidityReference {
	return z.humidityReference
}

func (z ZoneHumidity) Zone() string {
	return z.zone
}

func (z ZoneHumidity) CurrentHumidity() float64 {
	return z.currentHumidity
}

func (z ZoneHumidity) Measured() bool {
	return z.measured
}

func NewZoneHumidity(zone string, currentHumidity float64, humidityReference *HumidityReference) *ZoneHumidity {
	return &ZoneHumidity{zone: zone, currentHumidity: currentHumidity, measured: true, humidityReference: humidityReference}
}

func NewUnmeasuredZoneHumidity(zone string, humidityReference *HumidityReference) *ZoneHumidity {
	return &ZoneHumidity{zone: zone, humidityReference: humidityReference}
}
