package util

import (
	"github.com/paulbellamy/ratecounter"
	"time"
)

var successTPS *ratecounter.RateCounter
var failTPS *ratecounter.RateCounter

// LogSuccess increments success count
func LogSuccess() {
	if successTPS == nil {
		successTPS = ratecounter.NewRateCounter(1 * time.Second)
	}
	successTPS.Incr(1)
}

// LogFail increments fail count
func LogFail() {
	if failTPS == nil {
		failTPS = ratecounter.NewRateCounter(1 * time.Second)
	}
	failTPS.Incr(1)
}

// GetSuccessTPS returns the success tps
func GetSuccessTPS() int64 {
	if successTPS == nil {
		return 0
	}
	return successTPS.Rate()
}

// GetFailTPS returns the success tps
func GetFailTPS() int64 {
	if failTPS == nil {
		return 0
	}
	return failTPS.Rate()
}
