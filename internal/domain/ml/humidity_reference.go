package ml

type HumidityReference struct {
	v40  float64
	v100 float64
}

func (h HumidityReference) V40() float64 {
	return h.v40
}

func (h HumidityReference) V100() float64 {
	return h.v100
}

func (h HumidityReference) voltageForPercentage(p float64) float64 {
	return h.v100 + (h.v40-h.v100)*((100-p)/60)
}

func (h HumidityReference) LowHumidity() float64 {
	return h.v40
}

func (h HumidityReference) HighHumidity() float64 {
	return h.voltageForPercentage(60)
}

func (h HumidityReference) NotWorkingFineLimit() float64 {
	return h.voltageForPercentage(15)
}

func (h HumidityReference) IsLow(v float64) bool {
	return v >= h.LowHumidity() && v <= h.NotWorkingFineLimit()
}

func (h HumidityReference) IsHigh(v float64) bool {
	return v <= h.HighHumidity()
}

func (h HumidityReference) InRange(v float64) bool {
	return v <= h.LowHumidity() && v >= h.HighHumidity()
}

func (h HumidityReference) TargetMoistureVoltage() float64 {
	return h.voltageForPercentage(90)
}

func NewHumidityReference(v40, v100 float64) *HumidityReference {
	return &HumidityReference{
		v40:  v40,
		v100: v100,
	}
}
