package ucp

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Ucp Operations
const (
	AlertOp               = "31"
	SubmitShortMessageOp  = "51"
	DeliverShortMessageOp = "52"
	DeliverNotificationOp = "53"
	SessionManagementOp   = "60"
)

// ucp start and terminator
const (
	STX = 2
	ETX = 3
)

// PDU type
const (
	OperationType = "O"
	ResultType    = "R"
)

// PDU represents a single request
type PDU struct {
	TransRefNum string
	Len         string
	Type        string
	Operation   string
	Data        []string
	Checksum    string
}

// NewPDU creates an instance of PDU from raw bytes
func NewPDU(raw []byte) (*PDU, error) {
	log.Println("Deserializing: ", string(raw))

	if len(raw) == 0 {
		return nil, errors.New("Empty packet")
	}
	if raw[0] != STX {
		return nil, errors.New("Invalid STX")
	}
	if raw[len(raw)-1] != ETX {
		return nil, errors.New("Invalid ETX")
	}

	pdu := new(PDU)
	pdu.TransRefNum = string(raw[1:3])
	pdu.Len = string(raw[4:9])
	switch string(raw[10]) {
	case OperationType:
		pdu.Type = OperationType
	case ResultType:
		pdu.Type = ResultType
	default:
		return nil, errors.New("Invalid PDU type")
	}
	pdu.Operation = string(raw[12:14])
	pdu.Data = strings.Split(string(raw[15:len(raw)-4]), "/")
	pdu.Checksum = string(raw[len(raw)-3 : len(raw)-1])
	return pdu, nil
}

// Bytes returns the serialized PDU
func (pdu *PDU) Bytes() []byte {
	// compute len
	var length int
	length += len(pdu.TransRefNum) + 1
	length += len(pdu.Operation) + 1
	length += len(pdu.Type) + 1
	length += 6 // len(5) + /(1)
	for i := range pdu.Data {
		length += len(pdu.Data[i]) + 1
	}
	length += 2 // checksum

	var b bytes.Buffer

	b.WriteString(pdu.TransRefNum)
	b.WriteString("/")
	b.WriteString(fmt.Sprintf("%05d", length))
	b.WriteString("/")
	b.WriteString(pdu.Type)
	b.WriteString("/")
	b.WriteString(pdu.Operation)
	b.WriteString("/")
	for i := range pdu.Data {
		b.WriteString(pdu.Data[i])
		if i < len(pdu.Data)-1 {
			b.WriteString("/")
		}
	}
	b.WriteString("/")
	partial := []byte(b.String())
	chkSum := checkSum(partial)
	res := make([]byte, 0)
	res = append(res, STX)
	res = append(res, partial...)
	res = append(res, chkSum...)
	res = append(res, ETX)
	return res
}

// checkSum computes the checksum of the pdu
func checkSum(b []byte) []byte {
	var sum byte
	for _, i := range b {
		sum += i
	}
	mask := sum & 0xFF
	ck := strings.ToUpper(strconv.FormatInt(int64(mask), 16))
	chkSum := fmt.Sprintf("%02s", ck)
	log.Println("chkSum: ", chkSum)
	return []byte(chkSum)
}
