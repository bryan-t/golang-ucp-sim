package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	SubmitSMResponseTimeLow  int
	SubmitSMResponseTimeHigh int
	SubmitSMWindowMax        int
	DeliverSMWindowMax       int
	APIPort                  int
	UcpPort                  int
	MaxTPS                   int64
}

var instance *config

// GetConfig gets the config instance
func GetConfig() *config {
	if instance == nil {
		instance = new(config)
		instance.SubmitSMResponseTimeLow = 0
		instance.SubmitSMResponseTimeHigh = 0
		instance.SubmitSMWindowMax = 100
		instance.DeliverSMWindowMax = 99
		instance.APIPort = 8090
		instance.UcpPort = 8080
		instance.MaxTPS = 100

		readConfig()
	}
	return instance
}

func readConfig() {
	confDir := os.Getenv("UCP_SIM_CONF_DIR")
	log.Println("Config Dir: ", confDir)
	var err error
	if confDir == "" {
		confDir, err = os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current working dir. Please set UCP_SIM_CONF_DIR on your env.")
			return
		}
		os.Setenv("UCP_SIM_CONF_DIR", confDir)
	}

	confName := "ucp-sim-conf.json"
	confPath := fmt.Sprintf("%s%s%s", confDir, string(os.PathSeparator), confName)
	config := instance
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		configJson, _ := json.MarshalIndent(config, "", "\t")
		log.Println("Creating ", confPath)
		f, err := os.Create(confPath)
		if err != nil {
			log.Fatal("Encountered error while creating config file: ", err)
		}
		_, err = f.Write(configJson)
		f.Close()
		if err != nil {
			log.Fatal("Encountered error while writing config file: ", err)
			return
		}
	} else if err == nil {
		log.Println("Opening ", confPath)
		f, err := os.Open(confPath)
		if err != nil {
			log.Fatal("Encountered error while opening config file: ", err)
			return
		}
		confBytes, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal("Encountered error while reading config file: ", err)
			return
		}
		err = json.Unmarshal(confBytes, &config)
		if err != nil {
			log.Fatal("Encountered error while parsing config file: ", err)
			return
		}
	} else {
		log.Fatal("Encountered error while checking config file: ", err)
		return
	}
	if config.DeliverSMWindowMax > 99 {
		panic("DeliverSMWindowMax should be at most 99")
	}
	log.Printf("Using config: \n%+v\n", config)
}
