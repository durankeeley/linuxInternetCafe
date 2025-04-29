package utils

import (
	"time"
)

func MinutesToDuration(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}
