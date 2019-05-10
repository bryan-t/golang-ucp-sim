package ucpsvr

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/models"
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"github.com/bryan-t/golang-ucp-sim/ucpmock"
	"github.com/bryan-t/golang-ucp-sim/util"
	"log"
	"net"
	"time"
)

// UcpServer a server which processes incoming UCP requests
type UcpServer struct {
	listener    net.Listener
	clients     *ClientSlice
	deliverChan chan *models.DeliverSMReq
}

// ETX is the terminator for UCP packets
const ETX = 3

func NewUcpServer() *UcpServer {
	server := new(UcpServer)
	server.clients = NewClientSlice()
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
		log.Println("Got a new connection.")
		// NOTE: need to transfer functions from server to client
		go server.handleIncoming(client)
		go server.processDeliver(client)
	}

}

// Stop stops the UcpServer
func (server *UcpServer) Stop() {
	server.listener.Close()
	for _, client := range server.clients.GetClients() {
		(*(*client).conn).Close()
	}
}

func (server *UcpServer) handleIncoming(client *client) {
	reader := bufio.NewReader(*client.conn)
	server.clients.Append(client)
	defer server.clients.Remove(client)
	channel := make(chan *ucp.PDU, util.MaxWindowSize)
	go server.processIncomingViaChannel(client, channel)
	for {

		data, err := reader.ReadSlice(ETX)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			(*client.conn).Close()
			return
		}

		pdu, err := ucp.NewPDU(data)
		log.Println("Got PDU: ", pdu)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			(*client.conn).Close()
			return
		}
		config := util.GetConfig()
		max := config.SubmitSMWindowMax
		if max > util.MaxWindowSize {
			max = util.MaxWindowSize
		}

		if pdu.Operation != ucp.SubmitShortMessageOp {
			server.processIncoming(client, pdu)
			continue
		}

		if len(channel) >= max {
			log.Println("Max window reached")
			util.LogFail()
			res := ucp.NewSubmitSMResponse(pdu, true, "MAX WINDOW")
			resBytes := res.Bytes()
			_, err = (*client.conn).Write(resBytes)
			if err != nil {
				log.Println("Encountered error ", err.Error())
				(*client.conn).Close()
				return
			}
			continue
		}
		channel <- pdu
	}
}

func (server *UcpServer) processIncoming(client *client, pdu *ucp.PDU) {
	res, _ := ucpmock.ProcessIncoming(pdu)
	if pdu != nil && pdu.Type == ucp.ResultType && pdu.Operation == ucp.DeliverShortMessageOp {
		processDeliverSMResult(client, pdu)
		return
	}
	if res == nil {
		return
	}
	resBytes := res.Bytes()
	_, err := (*client.conn).Write(resBytes)
	if err != nil {
		log.Println("Encountered error ", err.Error())
		(*client.conn).Close()
		return
	}
}

func processDeliverSMResult(client *client, pdu *ucp.PDU) {
	client.removeFromWindow(pdu.TransRefNum)
}

func (server *UcpServer) processIncomingViaChannel(client *client, c chan *ucp.PDU) {
	for {
		pdu, ok := <-c
		if !ok {
			break
		}
		log.Println("Got a new request from channel")
		server.processIncoming(client, pdu)
	}
}

func (server *UcpServer) processDeliver(client *client) {
	for {
		if !client.windowHasVacantSlot() {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		log.Println("Client: Waiting for request.")
		req, _ := <-server.deliverChan
		log.Println("Got a new deliver request.")
		err := server.processDeliverReq(req, client)
		if err != nil {
			return
		}
	}
}

func (server *UcpServer) processDeliverReq(req *models.DeliverSMReq, client *client) error {
	log.Printf("Processing: %+v\n", req)
	pdu := ucp.NewDeliverSMPDU(req.Recipient, req.AccessCode, req.Message)
	client.putToWindow(pdu)
	log.Printf("Sending: %+v\n", pdu)
	_, err := (*client.conn).Write(pdu.Bytes())
	if err != nil {
		log.Println("Failed to send Deliver SM: ", err)
		(*client.conn).Close()
		server.deliverChan <- req
		return errors.New("Failed to deliver SM")
	}
	util.LogIncomingTPS()
	return nil
}
