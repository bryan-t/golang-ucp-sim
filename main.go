package main

import (
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

	config := util.GetConfig()
	var server = ucpsvr.NewUcpServer()
	go httpsvr.Start(config.APIPort, server.Deliver)
	server.Start(config.UcpPort)

}
