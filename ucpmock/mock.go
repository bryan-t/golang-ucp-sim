package ucpmock

import (
	"github.com/bryan-t/golang-ucp-sim/ucp"
	"log"
)

// ProcessIncoming is mock handling of incoming pdu requests
func ProcessIncoming(req *ucp.PDU) (*ucp.PDU, error) {
	switch req.Operation {
	case ucp.AlertOp:
	case ucp.SubmitShortMessageOp:
	case ucp.DeliverShortMessageOp:
	case ucp.DeliverNotificationOp:
	case ucp.SessionManagementOp:
		return ProcessSessionManagement(req)
	default:
		log.Println("Got unsupported op: ", req.Operation)
		return nil, nil
	}
	return nil, nil
}

// ProcessSessionManagement accepts all incoming auth request
func ProcessSessionManagement(req *ucp.PDU) (*ucp.PDU, error) {
	res := new(ucp.PDU)
	res.TransRefNum = req.TransRefNum
	res.Type = ucp.ResultType
	res.Operation = req.Operation
	res.Data = append(res.Data, "A")
	res.Data = append(res.Data, "BIND AUTHENTICATED")
	return res, nil
}
