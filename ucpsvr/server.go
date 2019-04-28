package ucpsvr

import (
	"bufio"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/common"
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"github.com/bryan-t/golang-ucp-sim/ucpmock"
	"github.com/bryan-t/golang-ucp-sim/util"
	"log"
	"net"
)

// UcpServer a server which processes incoming UCP requests
type UcpServer struct {
	listener net.Listener
	conns    *common.ConnSlice
}

// ETX is the terminator for UCP packets
const ETX = 3

// NewUcpServer creates a new instance of UcpServer
func NewUcpServer() *UcpServer {
	server := new(UcpServer)
	server.conns = common.NewConnSlice()
	server.listener = nil
	return server
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

		server.conns.Append(conn)
		log.Println("Got a new connection.")
		go handleIncoming(conn)
	}

}

// Stop stops the UcpServer
func (server *UcpServer) Stop() {
	server.listener.Close()
	for _, conn := range server.conns.GetConns() {
		conn.Close()
	}
}

func handleIncoming(conn net.Conn) {
	reader := bufio.NewReader(conn)
	channel := make(chan *ucp.PDU, util.MaxWindowSize)
	go processIncomingViaChannel(conn, channel)
	for {

		data, err := reader.ReadSlice(ETX)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			conn.Close()
			return
		}

		pdu, err := ucp.NewPDU(data)
		log.Println("Got PDU: ", pdu)
		if err != nil {
			log.Println("Encountered error ", err.Error())
			conn.Close()
			return
		}
		config := util.GetConfig()
		max := config.SubmitSMWindowMax
		if max > util.MaxWindowSize {
			max = util.MaxWindowSize
		}

		if pdu.Operation != ucp.SubmitShortMessageOp {
			processIncoming(conn, pdu)
			continue
		}

		if len(channel) >= max {
			log.Println("Max window reached")
			util.LogFail()
			res := ucp.NewSubmitSMResponse(pdu, true, "MAX WINDOW")
			resBytes := res.Bytes()
			conn.Write(resBytes)
			continue
		}
		channel <- pdu
	}
}

func processIncoming(conn net.Conn, pdu *ucp.PDU) {
	res, _ := ucpmock.ProcessIncoming(pdu)
	if res == nil {
		return
	}
	resBytes := res.Bytes()
	conn.Write(resBytes)
}

func processIncomingViaChannel(conn net.Conn, c chan *ucp.PDU) {
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
		conn.Write(resBytes)
	}
}
