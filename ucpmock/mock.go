package ucpmock

import (
	"golang-ucp-sim/ucp"
	"golang-ucp-sim/util"
	"log"
	"math/rand"
	"time"
)

// ProcessIncoming is mock handling of incoming pdu requests
func ProcessIncoming(req *ucp.PDU) (*ucp.PDU, error) {
	switch req.Operation {
	case ucp.AlertOp:
		return ProcessAlertOp(req)
	case ucp.SubmitShortMessageOp:
		return ProcessSubmitShortMessageOp(req)
	case ucp.DeliverShortMessageOp:
	case ucp.DeliverNotificationOp:
	case ucp.SessionManagementOp:
		return ProcessSessionManagementOp(req)
	default:
		log.Println("Got unsupported op: ", req.Operation)
		return nil, nil
	}
	return nil, nil
}

// ProcessSessionManagementOp accepts all incoming auth request
func ProcessSessionManagementOp(req *ucp.PDU) (*ucp.PDU, error) {
	res := new(ucp.PDU)
	res.TransRefNum = req.TransRefNum
	res.Type = ucp.ResultType
	res.Operation = req.Operation
	res.Data = append(res.Data, "A")
	res.Data = append(res.Data, "BIND AUTHENTICATED")
	return res, nil
}

// ProcessAlertOp always returns success for alert operations
func ProcessAlertOp(req *ucp.PDU) (*ucp.PDU, error) {
	// TODO: might be good to create NewPDU func for each type. This way, pdu.Data could be abstracted
	// away
	res := new(ucp.PDU)
	res.TransRefNum = req.TransRefNum
	res.Type = ucp.ResultType
	res.Operation = req.Operation
	res.Data = append(res.Data, "A")
	res.Data = append(res.Data, "0000")
	return res, nil
}

// ProcessSubmitShortMessageOp returns success for all submit sm request. Doesn't support delivery receipt yet
func ProcessSubmitShortMessageOp(req *ucp.PDU) (*ucp.PDU, error) {
	msg, _ := req.GetMessage()
	sender, _ := req.GetSender()
	recipient, _ := req.GetRecipient()
	log.Printf("Received a message '%s' from sender '%s' to recipient '%s' \n", msg, sender, recipient)
	res := ucp.NewSubmitSMResponse(req, true, "")

	config := util.GetConfig()
	diff := config.SubmitSMResponseTimeHigh - config.SubmitSMResponseTimeLow
	sleep := config.SubmitSMResponseTimeLow
	if diff != 0 {
		sleep = rand.Int()%diff + config.SubmitSMResponseTimeLow
	}
	if sleep > 0 {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	return res, nil
}
