package common

import (
	"log"
	"net"
	"sync"
)

type ConnSlice struct {
	conns []*net.Conn
	mutex sync.Mutex
}

func NewConnSlice() *ConnSlice {
	slice := new(ConnSlice)
	slice.conns = make([]*net.Conn, 0, 10)
	slice.mutex = sync.Mutex{}
	return slice
}

func (slice *ConnSlice) Append(conn *net.Conn) {
	slice.mutex.Lock()
	slice.conns = append(slice.conns, conn)
	slice.mutex.Unlock()
}

func (slice *ConnSlice) Remove(conn *net.Conn) {
	slice.mutex.Lock()
	for i := range slice.conns {
		if (*slice.conns[i]).RemoteAddr() != (*conn).RemoteAddr() {
			continue
		}
		slice.conns = append(slice.conns[:i], slice.conns[i+1:]...)
	}
	slice.mutex.Unlock()
}

func (slice *ConnSlice) GetConns() []*net.Conn {
	slice.mutex.Lock()
	conns := slice.conns
	slice.mutex.Unlock()
	return conns
}

// GetConn returns the connection at the specified index. If the index exceeds the size, returns the first one
func (slice *ConnSlice) GetConn(i int) (*net.Conn, int) {
	slice.mutex.Lock()
	defer slice.mutex.Unlock()

	if len(slice.conns) == 0 {
		return nil, -1
	}

	if i >= len(slice.conns) {
		log.Println("Index is greater than length")
		return slice.conns[0], 0
	}
	return slice.conns[i], i
}
