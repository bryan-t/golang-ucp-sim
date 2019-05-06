package models

type DeliverSMReq struct {
	AccessCode string
	Recipient  string
	Message    string
}

type DeliverSMReqBulk struct {
	Requests []DeliverSMReq
}
