package ucpsvr

import (
	"bufio"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/common"
	"github.com/bryan-t/golang-ucp-sim/models"
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"github.com/bryan-t/golang-ucp-sim/ucpmock"
	"github.com/bryan-t/golang-ucp-sim/util"
	"log"
	"net"
)

// UcpServer a server which processes incoming UCP requests
type UcpServer struct {
	listener    net.Listener
	conns       *common.ConnSlice
	deliverChan chan *models.DeliverSMReq
}

// ETX is the terminator for UCP packets
const ETX = 3

func NewUcpServer() *UcpServer {
	server := new(UcpServer)
	server.conns = common.NewConnSlice()
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
	go server.processDeliver()
	for {
		conn, listenErr := server.listener.Accept()
		if err != nil {
			log.Println(listenErr)
		}

		log.Println("Got a new connection.")
		go server.handleIncoming(&conn)
	}

}

// Stop stops the UcpServer
func (server *UcpServer) Stop() {
	server.listener.Close()
	for _, conn := range server.conns.GetConns() {
		(*conn).Close()
	}
}

func (server *UcpServer) handleIncoming(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	server.conns.Append(conn)
	defer server.conns.Remove(conn)
	channel := make(chan *ucp.PDU, util.MaxWindowSize)
	go server.processIncomingViaChannel(conn, channel)
	for {

		data, err := reader.ReadSlice(ETX)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			(*conn).Close()
			return
		}

		pdu, err := ucp.NewPDU(data)
		log.Println("Got PDU: ", pdu)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			(*conn).Close()
			return
		}
		config := util.GetConfig()
		max := config.SubmitSMWindowMax
		if max > util.MaxWindowSize {
			max = util.MaxWindowSize
		}

		if pdu.Operation != ucp.SubmitShortMessageOp {
			server.processIncoming(conn, pdu)
			continue
		}

		if len(channel) >= max {
			log.Println("Max window reached")
			util.LogFail()
			res := ucp.NewSubmitSMResponse(pdu, true, "MAX WINDOW")
			resBytes := res.Bytes()
			_, err = (*conn).Write(resBytes)
			if err != nil {
				(*conn).Close()
				return
			}
			continue
		}
		channel <- pdu
	}
}

func (server *UcpServer) processIncoming(conn *net.Conn, pdu *ucp.PDU) {
	res, _ := ucpmock.ProcessIncoming(pdu)
	if res == nil {
		return
	}
	resBytes := res.Bytes()
	_, err := (*conn).Write(resBytes)
	if err != nil {
		(*conn).Close()
		return
	}
}

func (server *UcpServer) processIncomingViaChannel(conn *net.Conn, c chan *ucp.PDU) {
	for {
		pdu, ok := <-c
		if !ok {
			break
		}
		log.Println("Got a new request from channel")
		res, _ := ucpmock.ProcessIncoming(pdu)
		if res == nil {
			continue
		}
		resBytes := res.Bytes()
		_, err := (*conn).Write(resBytes)
		if err != nil {
			(*conn).Close()
			return
		}
	}
}

func (server *UcpServer) processDeliver() {
	i := 0
	for {
		req, _ := <-server.deliverChan
		log.Println("Got a new deliver request. Spawning a routine")
		var conn *net.Conn
		conn, i = server.conns.GetConn(i)
		if conn == nil {
			log.Println("Got no connection. Discarding Deliver SM")
			continue
		}
		i++
		go server.processDeliverReq(req, conn)
	}
}

func (server *UcpServer) processDeliverReq(req *models.DeliverSMReq, conn *net.Conn) {
	log.Printf("Processing: %+v\n", req)
	pdu := ucp.NewDeliverSMPDU(req.Recipient, req.AccessCode, req.Message)
	log.Printf("Sending: %+v\n", pdu)
	_, err := (*conn).Write(pdu.Bytes())
	if err != nil {
		log.Println("Failed to send Deliver SM: ", err)
		(*conn).Close()
		server.deliverChan <- req
	}
	util.LogIncomingTPS()
}
