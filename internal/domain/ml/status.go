package ml

type Status struct {
	active  bool
	raining bool
}

func (s Status) Raining() bool {
	return s.raining
}

func (s Status) Active() bool {
	return s.active
}

func NewStatus(active, raining bool) *Status {
	return &Status{active: active, raining: raining}
}
