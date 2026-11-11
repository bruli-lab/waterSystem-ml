package app_test

import (
	"time"
)

const (
	bonsaiBigZone   = "bonsai big"
	bonsaiSmallZone = "bonsai small"
)

var fixedNow = time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

func clock() time.Time {
	return fixedNow
}

const (
	humidity100    = 1.2
	humidity40     = 1.8
	lowHumidity    = 1.85
	mediumHumidity = 1.7
)

func dayTimeFunc() time.Time {
	return time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC)
}
