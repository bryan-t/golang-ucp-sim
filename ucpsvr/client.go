package ucpsvr

import (
	"bufio"
	"errors"
	"fmt"
	"golang-ucp-sim/models"
	"golang-ucp-sim/ucp"
	"golang-ucp-sim/ucpmock"
	"golang-ucp-sim/util"
	"log"
	"net"
	"sync"
	"time"
)

const transRefMax = 99

type client struct {
	id                     int
	windowMutex            sync.Mutex
	conn                   *net.Conn
	currentDeliverTransRef int
	deliverWindow          map[string]*ucp.PDU
	deliverChan            chan *models.DeliverSMReq
	stop                   bool
	globalStop             *bool
	receivedFile           *log.Logger
	outgoingFile           *log.Logger
}

func (c *client) windowHasVacantSlot() bool {
	c.windowMutex.Lock()
	defer c.windowMutex.Unlock()
	log.Printf("Client id '%d' - Max window size: '%d' | Current size: '%d' ",
		c.id, util.GetConfig().DeliverSMWindowMax, len(c.deliverWindow))
	return len(c.deliverWindow) < util.GetConfig().DeliverSMWindowMax
}

/*
func (c *client) reserveWindowSlot() string {
	c.windowMutex.Lock()
	defer c.windowMutex.Unlock()
	if len(c.deliverWindow) > util.GetConfig().DeliverSMWindowMax {
		return 0
	}

	for {
		c.currentDeliverTransRef = (c.currentDeliverTransRef)%transRefMax + 1
		idx = fmt.Sprintf("%02d", c.currentDeliverTransRef)
		if _, ok := c.deliverWindow[idx]; ok {
			continue
		}
		break
	}
}
*/

func (c *client) start(wg *sync.WaitGroup) {
	var cWG sync.WaitGroup
	wg.Add(1)
	go c.handleIncoming(&cWG)
	go c.processDeliver(&cWG)
	cWG.Wait()
	wg.Done()
}

func (c *client) putToWindow(pdu *ucp.PDU) string {
	c.windowMutex.Lock()
	defer c.windowMutex.Unlock()
	if len(c.deliverWindow) > util.GetConfig().DeliverSMWindowMax {
		return ""
	}
	idx := ""
	for {
		c.currentDeliverTransRef = (c.currentDeliverTransRef)%transRefMax + 1
		idx = fmt.Sprintf("%02d", c.currentDeliverTransRef)
		if _, ok := c.deliverWindow[idx]; ok {
			continue
		}
		break
	}
	c.deliverWindow[idx] = pdu
	pdu.TransRefNum = idx
	return idx
}

func (c *client) removeFromWindow(transRefNum string) {
	c.windowMutex.Lock()
	defer c.windowMutex.Unlock()
	log.Println("Removing: ", transRefNum)
	delete(c.deliverWindow, transRefNum)
}

func (client *client) handleIncoming(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	reader := bufio.NewReader(*client.conn)
	//server.clients.Append(client)
	//defer server.clients.Remove(client)
	channel := make(chan *ucp.PDU, util.MaxWindowSize)
	go client.processIncomingViaChannel(channel)
	for !client.stop && !*client.globalStop {

		data, err := reader.ReadSlice(ETX)
		if err != nil {
			log.Println(client.conn, "Read slice Encountered error ", data, err.Error())
			(*client.conn).Close()
			client.stop = true
			return
		}

		pdu, err := ucp.NewPDU(data)
		log.Printf("Client id: '%d' - Incoming PDU: %s\n", client.id, string(pdu.Bytes()))
		if err != nil {
			log.Println("New PDU Encountered error ", err.Error())
			(*client.conn).Close()
			client.stop = true
			return
		}
		config := util.GetConfig()
		max := config.SubmitSMWindowMax
		if max > util.MaxWindowSize {
			max = util.MaxWindowSize
		}

		if pdu.Operation != ucp.SubmitShortMessageOp {
			client.processIncoming(pdu)
			continue
		}

		if len(channel) >= max {
			log.Println("Max window reached")
			util.LogFail()
			res := ucp.NewSubmitSMResponse(pdu, true, "MAX WINDOW")
			resBytes := res.Bytes()
			_, err = (*client.conn).Write(resBytes)
			if err != nil {
				log.Println("Write Encountered error ", err.Error())
				(*client.conn).Close()
				client.stop = true
				return
			}
			continue
		}
		channel <- pdu
	}
	close(channel)
}
func (client *client) processIncomingViaChannel(c chan *ucp.PDU) {
	for {
		pdu, ok := <-c
		if !ok {
			break
		}
		log.Println("Got a new request from channel")
		client.processIncoming(pdu)
	}
}
func (client *client) processIncoming(pdu *ucp.PDU) {
	ack, _ := ucpmock.ProcessIncoming(pdu)
	if pdu != nil && pdu.Type == ucp.ResultType && pdu.Operation == ucp.DeliverShortMessageOp {
		processDeliverSMResult(client, pdu)
		return
	}
	if pdu != nil && pdu.Type == ucp.OperationType && pdu.Operation == ucp.SubmitShortMessageOp {
		util.LogIncomingTPS()
		recipient, _ := pdu.GetRecipient()
		now := time.Now().Format("2006-01-02 15:04:05")
		toWrite := now + "," + recipient
		client.receivedFile.Println(toWrite)

	}

	if ack == nil {
		return
	}
	ackBytes := ack.Bytes()
	_, err := (*client.conn).Write(ackBytes)
	if err != nil {
		log.Println("processIncoming Encountered error ", err.Error())
		(*client.conn).Close()
		client.stop = true
		return
	}
}
func (client *client) processDeliver(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for !client.stop && !*client.globalStop {
		if !client.windowHasVacantSlot() {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		select {
		case req := <-client.deliverChan:
			log.Println("Got a new deliver request.")
			err := client.processDeliverReq(req)
			if err != nil {
				return
			}
		default:

			time.Sleep(1 * time.Second)
		}
	}
}

func (client *client) processDeliverReq(req *models.DeliverSMReq) error {
	pdu := ucp.NewDeliverSMPDU(req.Recipient, req.AccessCode, req.Message)
	client.putToWindow(pdu)
	pduBytes := pdu.Bytes()

	// Get current TPS
	//
	hour, min, sec := time.Now().Clock()
	year, month, day := time.Now().Date()
	sleepTime := time.Until(time.Date(year, month, day, hour, min, sec+1, 0, time.Local))

	if util.GetSuccessTPS() >= util.GetMaxTPS() {
		log.Println("Max TPS reached. Sleeping for", sleepTime)
		time.Sleep(sleepTime)
	}
	log.Printf("Sending: %s\n", string(pduBytes))
	_, err := (*client.conn).Write(pduBytes)
	if err != nil {
		log.Println(client.conn, "Failed to send Deliver SM: ", err)
		(*client.conn).Close()
		client.deliverChan <- req
		client.stop = true
		return errors.New("Failed to deliver SM")
	}
	util.LogSuccess()
	recipient, _ := pdu.GetRecipient()
	sendTime := time.Now().Format("2006-01-02 15:04:05")
	toWrite := sendTime + "," + recipient
	client.outgoingFile.Println(toWrite)
	return nil
}
func processDeliverSMResult(client *client, pdu *ucp.PDU) {
	client.removeFromWindow(pdu.TransRefNum)

}
