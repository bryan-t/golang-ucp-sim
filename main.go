package main

import (
	"encoding/json"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/httpsvr"
	"github.com/bryan-t/golang-ucp-sim/ucpsvr"
	"github.com/bryan-t/golang-ucp-sim/util"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	silent := os.Getenv("UCP_SIM_SILENT")
	if silent == "true" {
		log.SetOutput(ioutil.Discard)
	}

	log.Println("Initializing the server...")
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
	config := util.GetConfig()
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
	log.Printf("Using config: \n%+v\n", config)

	var server = ucpsvr.NewUcpServer()
	go httpsvr.Start(config.APIPort, server.Deliver)
	server.Start(config.UcpPort)

}
