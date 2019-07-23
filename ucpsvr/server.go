package ucpsvr

import (
	"fmt"
	"golang-ucp-sim/models"
	"golang-ucp-sim/ucp"
	"log"
	"net"
	"sync"
)

// UcpServer a server which processes incoming UCP requests
type UcpServer struct {
	listener     net.Listener
	clients      []*client
	deliverChan  chan *models.DeliverSMReq
	wg           sync.WaitGroup
	globalStop   bool
	receivedFile *log.Logger
	outgoingFile *log.Logger
}

// ETX is the terminator for UCP packets
const ETX = 3

func NewUcpServer(received *log.Logger, outgoing *log.Logger) *UcpServer {
	server := new(UcpServer)
	//server.clients = NewClientSlice()
	server.listener = nil
	server.receivedFile = received
	server.outgoingFile = outgoing
	server.deliverChan = make(chan *models.DeliverSMReq, 3000)
	server.globalStop = false
	return server
}

// Deliver queues the deliver request
func (server *UcpServer) Deliver(req *models.DeliverSMReq) {
	server.deliverChan <- req
}

// Start listens on the specified port
func (server *UcpServer) Start(port int) error {
	log.Println("Opening connection")
	networkString := fmt.Sprintf(":%d", port)
	var err error
	server.listener, err = net.Listen("tcp", networkString)
	if err != nil {
		log.Println("Failed to open connection")
		return err
	}
	id := 0
	for {
		conn, listenErr := server.listener.Accept()
		if listenErr != nil {
			//			log.Println(listenErr)
			continue
		}
		log.Println("Got a new connection.")
		client := new(client)
		client.id = id
		client.conn = &conn
		client.currentDeliverTransRef = 0
		client.deliverWindow = make(map[string]*ucp.PDU)
		client.deliverChan = server.deliverChan
		client.receivedFile = server.receivedFile
		client.outgoingFile = server.outgoingFile
		client.globalStop = &server.globalStop

		client.start(&server.wg)
		server.clients = append(server.clients, client)
		id += 1

		// TODO: proper stopping of clients
	}

}

// Stop stops the UcpServer
func (server *UcpServer) Stop() {
	log.Println("Stopping server")
	server.listener.Close()
	server.globalStop = true
	log.Println("Stopped server. Waiting for connections to be closed...")
	server.wg.Wait()
	log.Println("Closed all connections...")
	for _, client := range server.clients {
		(*(*client).conn).Close()
	}
	// TODO: closing of connections
}
