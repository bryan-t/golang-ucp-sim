package ucpsvr

import (
	"log"
	"sync"
)

type ClientSlice struct {
	clients []*client
	mutex   sync.Mutex
}

func NewClientSlice() *ClientSlice {
	slice := new(ClientSlice)
	slice.clients = make([]*client, 0, 10)
	slice.mutex = sync.Mutex{}
	return slice
}

func (slice *ClientSlice) Append(client *client) {
	slice.mutex.Lock()
	slice.clients = append(slice.clients, client)
	slice.mutex.Unlock()
}

func (slice *ClientSlice) Remove(client *client) {
	slice.mutex.Lock()
	for i := range slice.clients {
		if (*slice.clients[i].conn).RemoteAddr() != (*client.conn).RemoteAddr() {
			continue
		}
		slice.clients = append(slice.clients[:i], slice.clients[i+1:]...)
	}
	slice.mutex.Unlock()
}

func (slice *ClientSlice) GetClients() []*client {
	slice.mutex.Lock()
	clients := slice.clients
	slice.mutex.Unlock()
	return clients
}

// GetClient returns the connection at the specified index. If the index exceeds the size, returns the first one
func (slice *ClientSlice) GetClient(i int) (*client, int) {
	slice.mutex.Lock()
	defer slice.mutex.Unlock()

	if len(slice.clients) == 0 {
		return nil, -1
	}

	if i >= len(slice.clients) {
		log.Println("Index is greater than length")
		return slice.clients[0], 0
	}
	return slice.clients[i], i
}
