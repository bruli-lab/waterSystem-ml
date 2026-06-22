package ml

type ZoneHumidity struct {
	zone              string
	currentHumidity   float64
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

func NewZoneHumidity(zone string, currentHumidity float64, humidityReference *HumidityReference) *ZoneHumidity {
	return &ZoneHumidity{zone: zone, currentHumidity: currentHumidity, humidityReference: humidityReference}
}
