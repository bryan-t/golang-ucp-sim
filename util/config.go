package util

type config struct {
	SubmitSMResponseTimeLow  int
	SubmitSMResponseTimeHigh int
	SubmitSMWindowMax        int
}

var instance *config

func GetConfig() *config {
	if instance == nil {
		instance = new(config)
		instance.SubmitSMResponseTimeLow = 0
		instance.SubmitSMResponseTimeHigh = 0
		instance.SubmitSMWindowMax = 100
	}
	return instance
}
