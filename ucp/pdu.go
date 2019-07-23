package ucp

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	//	"github.com/go-gsm/charset"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
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

func (p PDU) String() string {
	return fmt.Sprintf("(TransRefNum: '%s', Len: '%s', Type: '%s', Operation: '%s', Data: '%v', Checksum: '%s')",
		p.TransRefNum, p.Len, p.Type, p.Operation, p.Data, p.Checksum)
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

// GetMessage returns the message when the operation is SubmitShortMessageOp
func (pdu *PDU) GetMessage() (string, error) {
	if pdu.Operation != SubmitShortMessageOp {
		return "", errors.New("Operation is not submit short message")
	}
	const (
		AMsg = "3"
		TD   = "4"
	)
	var msg string

	if pdu.Data[18] == AMsg {
		return decodeIRA([]byte(pdu.Data[20])), nil
	}
	if pdu.Data[18] == TD {
		decoded, err := hex.DecodeString(string(pdu.Data[20]))
		if err != nil {
			log.Println(err)
		}
		utf16, err := DecodeUcs2(decoded)
		if err != nil {
			log.Println(err)
		}
		msg = utf16
	}

	return msg, nil
}

// GetSender returns the sender of the message
func (pdu *PDU) GetSender() (string, error) {
	if pdu.Operation != SubmitShortMessageOp && pdu.Operation != DeliverShortMessageOp {
		err := fmt.Errorf("Operation %s does not have sender field", pdu.Operation)
		log.Println(err)
		return "", err
	}

	return pdu.Data[1], nil
}

// GetRecipient returns the recipient of message
func (pdu *PDU) GetRecipient() (string, error) {
	if pdu.Operation != SubmitShortMessageOp && pdu.Operation != DeliverShortMessageOp {
		err := fmt.Errorf("Operation %s does not have recipient field", pdu.Operation)
		log.Println(err)
		return "", err
	}
	return pdu.Data[0], nil
}

// DecodeUcs2 decodes the given UCS2 (UTF-16) octet data into a UTF-8 encoded string.
func DecodeUcs2(octets []byte) (str string, err error) {
	if len(octets)%2 != 0 {
		err = errors.New("DecodeUcs2: Uneven number of octets")
		return
	}
	buf := make([]uint16, 0, len(octets)/2)
	for i := 0; i < len(octets); i += 2 {
		buf = append(buf, uint16(octets[i])<<8|uint16(octets[i+1]))
	}
	runes := utf16.Decode(buf)
	return string(runes), nil
}

// EncodeUcs2 encodes the given UTF-8 text into UCS2 (UTF-16) encoding and returns the produced octets.
func EncodeUcs2(str string) []byte {
	buf := utf16.Encode([]rune(str))
	octets := make([]byte, 0, len(buf)*2)
	for _, n := range buf {
		octets = append(octets, byte(n&0xFF00>>8), byte(n&0x00FF))
	}
	return octets
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
	return []byte(chkSum)
}

// NewSubmitSMResponse creates a new response PDU for submit SM
func NewSubmitSMResponse(req *PDU, success bool, err string) *PDU {
	res := new(PDU)
	res.TransRefNum = req.TransRefNum
	res.Type = ResultType
	res.Operation = req.Operation
	if success {
		res.Data = append(res.Data, "A")
	} else {
		res.Data = append(res.Data, "N")
	}
	res.Data = append(res.Data, err)
	return res
}

// NewDeliverSMPDU creates a new deliver sm PDU
func NewDeliverSMPDU(recipient string, sender string, message string) *PDU {
	res := new(PDU)
	res.Type = OperationType
	res.Operation = DeliverShortMessageOp
	//	msgUcs := EncodeUcs2(message)

	msgIra := make([]byte, hex.EncodedLen(len(message)))
	hex.Encode(msgIra, []byte(message))
	res.Data = []string{
		recipient,
		sender,
		"", // AC
		"", // NRq
		"", // NAdC
		"", // NT
		"", // NPID
		"", // LRq
		"", // LRAd
		"", // LPID
		"", // DD
		"", // DDT
		"", // VP
		"", // RPID
		time.Now().Format("020106150405"),
		"",  // Dst
		"",  // Rsn
		"",  // DSCTS
		"3", // MT
		"",  // NB
		string(msgIra),
		"",       // MMS
		"",       // PR
		"",       // DCs
		"",       // MCLs
		"",       // RPI
		"",       // CPg
		"",       // RPLy
		"",       // OTOA
		"",       // HPLMN
		"020108", // xser
		"",       // RES4
		"",       // RES5
	}
	return res
}
