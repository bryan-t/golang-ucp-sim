package util

import (
	"sync"
	"time"
)

// RateCounter is a per second counter
//
type RateCounter struct {
	lock    sync.Mutex
	counter int64
	second  int
}

func (r *RateCounter) Incr() {
	r.lock.Lock()
	defer r.lock.Unlock()
	now := time.Now().Second()
	if now != r.second {
		r.counter = 0
		r.second = now
	}
	r.counter += 1
}

func (r *RateCounter) Rate() int64 {
	r.lock.Lock()
	defer r.lock.Unlock()
	now := time.Now().Second()
	if now != r.second {
		r.counter = 0
		r.second = now
	}

	return r.counter
}

var maxTPSMutex sync.Mutex
var maxTPS int64
var successTPS *RateCounter
var failTPS *RateCounter
var incomingTPS *RateCounter

func NewSuccessTPSCounter() {
	successTPS = &RateCounter{}
}
func NewIncomingTPSCounter() {
	incomingTPS = &RateCounter{}
}
func NewFailTPSCounter() {
	failTPS = &RateCounter{}
}

func GetMaxTPS() int64 {
	return maxTPS
}

func UpdateMaxTPS(newMaxTPS int64) {
	maxTPSMutex.Lock()
	defer maxTPSMutex.Unlock()
	maxTPS = newMaxTPS
	return
}

// LogIncomingTPS increments incoming TPS
func LogIncomingTPS() {
	incomingTPS.Incr()
}

// LogSuccess increments success count
func LogSuccess() {
	successTPS.Incr()
}

// LogFail increments fail count
func LogFail() {
	failTPS.Incr()
}

// GetSuccessTPS returns the success tps
func GetSuccessTPS() int64 {
	return successTPS.Rate()
}

// GetFailTPS returns the success tps
func GetFailTPS() int64 {
	return failTPS.Rate()
}

// GetIncomingTPS returns the incoming tps
func GetIncomingTPS() int64 {
	return incomingTPS.Rate()
}
