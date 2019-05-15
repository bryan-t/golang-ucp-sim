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
	"sync"
	"time"
)

const transRefMax = 99

type client struct {
	mutex                  sync.Mutex
	conn                   *net.Conn
	currentDeliverTransRef int
	deliverWindow          map[string]*ucp.PDU
	deliverChan            chan *models.DeliverSMReq
	stop                   *bool
}

func (c *client) windowHasVacantSlot() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Println("Current size: ", len(c.deliverWindow))
	log.Println("DeliverSMWindowMax: ", util.GetConfig().DeliverSMWindowMax)
	return len(c.deliverWindow) < util.GetConfig().DeliverSMWindowMax
}

/*
func (c *client) reserveWindowSlot() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	log.Println("Closing client..")
}

func (c *client) putToWindow(pdu *ucp.PDU) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.deliverWindow) > util.GetConfig().DeliverSMWindowMax {
		return ""
	}
	idx := ""
	for {
		c.currentDeliverTransRef = (c.currentDeliverTransRef)%transRefMax + 1
		idx = fmt.Sprintf("%02d", c.currentDeliverTransRef)
		log.Println("Trying index: ", idx)
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
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	for !*client.stop {

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
				log.Println("Encountered error ", err.Error())
				(*client.conn).Close()
				return
			}
			continue
		}
		channel <- pdu
	}
	close(channel)
}

func (client *client) processIncoming(pdu *ucp.PDU) {
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

func (client *client) processDeliver(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for !*client.stop {
		if !client.windowHasVacantSlot() {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		log.Println("Client: Waiting for request.")
		req, _ := <-client.deliverChan
		log.Println("Got a new deliver request.")
		err := client.processDeliverReq(req)
		if err != nil {
			return
		}
	}
}

func (client *client) processDeliverReq(req *models.DeliverSMReq) error {
	log.Printf("Processing: %+v\n", req)
	pdu := ucp.NewDeliverSMPDU(req.Recipient, req.AccessCode, req.Message)
	client.putToWindow(pdu)
	log.Printf("Sending: %+v\n", pdu)
	_, err := (*client.conn).Write(pdu.Bytes())
	if err != nil {
		log.Println("Failed to send Deliver SM: ", err)
		(*client.conn).Close()
		client.deliverChan <- req
		return errors.New("Failed to deliver SM")
	}
	util.LogIncomingTPS()
	return nil
}
