package ucpsvr

import (
	//"bufio"
	//"errors"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/models"
	"github.com/bryan-t/golang-ucp-sim/ucp"
	//"github.com/bryan-t/golang-ucp-sim/ucpmock"
	//"github.com/bryan-t/golang-ucp-sim/util"
	"log"
	"net"
	"sync"
	//"time"
)

// UcpServer a server which processes incoming UCP requests
type UcpServer struct {
	listener net.Listener
	//clients     *ClientSlice
	deliverChan chan *models.DeliverSMReq
	wg          sync.WaitGroup
}

// ETX is the terminator for UCP packets
const ETX = 3

func NewUcpServer() *UcpServer {
	server := new(UcpServer)
	//server.clients = NewClientSlice()
	server.listener = nil
	server.deliverChan = make(chan *models.DeliverSMReq, 100)
	return server
}

// Deliver queues the deliver request
func (server *UcpServer) Deliver(req *models.DeliverSMReq) {
	server.deliverChan <- req
}

// Start listens on the specified port
func (server *UcpServer) Start(port int) error {
	networkString := fmt.Sprintf(":%d", port)
	var err error
	server.listener, err = net.Listen("tcp", networkString)

	if err != nil {
		return err
	}
	for {
		conn, listenErr := server.listener.Accept()
		if err != nil {
			log.Println(listenErr)
		}
		client := new(client)
		client.conn = &conn
		client.currentDeliverTransRef = 0
		client.deliverWindow = make(map[string]*ucp.PDU)
		client.deliverChan = server.deliverChan
		log.Println("Got a new connection.")

		client.start()

		// TODO: proper stopping of clients
	}

}

// Stop stops the UcpServer
func (server *UcpServer) Stop() {
	server.listener.Close()
	/*for _, client := range server.clients.GetClients() {
		(*(*client).conn).Close()
	}*/
	// TODO: closing of connections
}
