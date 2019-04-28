package main

import (
	"github.com/bryan-t/golang-ucp-sim/httpsvr"
	"github.com/bryan-t/golang-ucp-sim/ucpsvr"
	"io/ioutil"
	"log"
)

func main() {
	log.SetOutput(ioutil.Discard)
	log.Println("Initializing the server...")
	go httpsvr.Start()
	var server = ucpsvr.NewUcpServer()
	server.Start(8080)

}
