package ml

type Status struct {
	active      bool
	raining     bool
	temperature float64
}

func (s Status) Raining() bool {
	return s.raining
}

func (s Status) Active() bool {
	return s.active
}

func (s Status) Temperature() float64 {
	return s.temperature
}

func NewStatus(active, raining bool, temperature float64) *Status {
	return &Status{active: active, raining: raining, temperature: temperature}
}
