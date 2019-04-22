package main

import (
	"github.com/bryan-t/golang-ucp-sim/ucpsvr"
	"log"
)

func main() {
	log.Println("Initializing the server...")
	var server = ucpsvr.NewUcpServer()
	server.Start(8080)

}
