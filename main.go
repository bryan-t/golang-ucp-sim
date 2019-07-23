package main

import (
	"context"
	"flag"
	"golang-ucp-sim/file"
	"golang-ucp-sim/httpsvr"
	"golang-ucp-sim/signal"
	"golang-ucp-sim/ucpsvr"
	"golang-ucp-sim/util"
	"io/ioutil"
	"log"
	"os"
	"syscall"
)

func main() {
	var outgoingFileName string
	var receivedFileName string
	flag.StringVar(&outgoingFileName, "o", "", "Outgoing log file")
	flag.StringVar(&receivedFileName, "r", "", "Received log file")
	flag.Parse()
	if outgoingFileName == "" {
		log.Fatal("Outgoing file argument missing.\n")
	}
	if receivedFileName == "" {
		log.Fatal("Received file argument missing.\n")
	}

	util.NewSuccessTPSCounter()
	util.NewIncomingTPSCounter()
	util.NewFailTPSCounter()
	// Open outgoing file
	//
	outgoingFile, err := file.NewFile(outgoingFileName)
	log.Printf("Received file %s", outgoingFileName)
	if err != nil {
		log.Fatalf("Failed to open file with error: %s.\n", err.Error())
	}
	defer outgoingFile.Close()
	outgoingLogger := outgoingFile.Logger()
	// Open recieve file
	//
	receiveFile, err := file.NewFile(receivedFileName)
	log.Printf("Received file %s", receivedFileName)
	if err != nil {
		log.Fatalf("Failed to open file with error: %s.\n", err.Error())
	}
	defer receiveFile.Close()
	receivedLogger := receiveFile.Logger()

	silent := os.Getenv("UCP_SIM_SILENT")
	if silent == "true" {
		log.SetOutput(ioutil.Discard)
	}

	log.Println("Initializing the server...")

	config := util.GetConfig()
	util.UpdateMaxTPS(config.MaxTPS)
	server := ucpsvr.NewUcpServer(receivedLogger, outgoingLogger)

	httpserver := httpsvr.NewServer(config.APIPort, server.Deliver)
	// Add signal handling
	logsRotate := func() {
		outgoingFile.Rotate()
		receiveFile.Rotate()
	}
	stopFn := func() {
		server.Stop()
		httpserver.Shutdown(context.TODO())
	}
	signal.Handle(syscall.SIGHUP, logsRotate)
	signal.Handle(syscall.SIGTERM, stopFn)
	// Start signalling
	//
	go signal.Listen()

	go server.Start(config.UcpPort)
	go log.Fatal(httpserver.ListenAndServe())

}
