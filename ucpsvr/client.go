package ucpsvr

import (
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"net"
)

type client struct {
	conn                   *net.Conn
	currentDeliverTransRef int
	deliverWindow          map[int]*ucp.PDU
}
