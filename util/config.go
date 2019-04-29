package util

type config struct {
	SubmitSMResponseTimeLow  int
	SubmitSMResponseTimeHigh int
	SubmitSMWindowMax        int
	APIPort                  int
	UcpPort                  int
}

var instance *config

// GetConfig gets the config instance
func GetConfig() *config {
	if instance == nil {
		instance = new(config)
		instance.SubmitSMResponseTimeLow = 0
		instance.SubmitSMResponseTimeHigh = 0
		instance.SubmitSMWindowMax = 100
		instance.APIPort = 8090
		instance.UcpPort = 8080

	}
	return instance
}
