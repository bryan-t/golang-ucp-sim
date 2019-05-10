package ucpsvr

import (
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"github.com/bryan-t/golang-ucp-sim/util"
	"log"
	"net"
	"sync"
)

const transRefMax = 99

type client struct {
	mutex                  sync.Mutex
	conn                   *net.Conn
	currentDeliverTransRef int
	deliverWindow          map[string]*ucp.PDU
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
