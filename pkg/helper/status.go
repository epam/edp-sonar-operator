package helper

import (
	"math"
	"time"
)

const (
	StatusOK = "OK"
	TimeOut  = 10 * time.Second
)

type FailureCountable interface {
	GetFailureCount() int64
	SetFailureCount(count int64)
}

type StatusValue interface {
	GetStatus() string
	SetStatus(val string)
}

type StatusValueFailureCountable interface {
	FailureCountable
	StatusValue
}

func SetFailureCount(fc FailureCountable) time.Duration {
	failures := fc.GetFailureCount()
	timeout := getTimeout(failures, TimeOut)
	failures += 1
	fc.SetFailureCount(failures)

	return timeout
}

func getTimeout(factor int64, baseDuration time.Duration) time.Duration {
	return time.Duration(float64(baseDuration) * math.Pow(math.E, float64(factor+1)))
}

func SetSuccessStatus(el StatusValueFailureCountable) {
	el.SetStatus(StatusOK)
	el.SetFailureCount(0)
}
