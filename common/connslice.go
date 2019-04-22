package common

import (
	"net"
	"sync"
)

type ConnSlice struct {
	conns []net.Conn
	mutex sync.Mutex
}

func NewConnSlice() *ConnSlice {
	slice := new(ConnSlice)
	slice.conns = make([]net.Conn, 0, 10)
	slice.mutex = sync.Mutex{}
	return slice
}

func (slice *ConnSlice) Append(conn net.Conn) {
	slice.mutex.Lock()
	slice.conns = append(slice.conns, conn)
	slice.mutex.Unlock()
}

func (slice *ConnSlice) Remove(conn net.Conn) {
	slice.mutex.Lock()
	for i := range slice.conns {
		if slice.conns[i].RemoteAddr() != conn.RemoteAddr() {
			continue
		}
		slice.conns = append(slice.conns[:i], slice.conns[i+1:]...)
	}
	slice.mutex.Unlock()
}

func (slice *ConnSlice) GetConns() []net.Conn {
	slice.mutex.Lock()
	conns := slice.conns
	slice.mutex.Unlock()
	return conns
}
