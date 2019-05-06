package common

import (
	"github.com/bryan-t/golang-ucp-sim/models"
	"sync"
)

type UcpReqSlice struct {
	objs  []*models.DeliverSMReq
	mutex sync.Mutex
}

func (slice *UcpReqSlice) Append(obj *models.DeliverSMReq) {
	slice.mutex.Lock()
	slice.objs = append(slice.objs, obj)
	slice.mutex.Unlock()
}

func (slice *UcpReqSlice) Get() *models.DeliverSMReq {
	slice.mutex.Lock()
	defer slice.mutex.Unlock()
	if len(slice.objs) == 0 {
		return nil
	}
	req := slice.objs[0]
	slice.objs = slice.objs[1:]
	return req

}
