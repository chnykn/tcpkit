package tcp

import (
	"log"
	"sync"
)

//TCPConnBucket 用来存放和管理TCPConn连接
type TCPConnBucket struct {
	m  map[string]*TCPConn
	mu *sync.RWMutex
}

func NewTCPConnBucket() *TCPConnBucket {
	tcb := &TCPConnBucket{
		m:  make(map[string]*TCPConn),
		mu: new(sync.RWMutex),
	}
	return tcb
}

func (self *TCPConnBucket) Put(id string, c *TCPConn) {
	log.Println("TCPConnBucket PUT id =", id)

	self.mu.Lock()
	if conn, ok := self.m[id]; ok {
		conn.Close()
	}
	self.m[id] = c
	self.mu.Unlock()
}

func (self *TCPConnBucket) Get(id string) *TCPConn {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if conn, ok := self.m[id]; ok {
		return conn
	}
	return nil
}

func (self *TCPConnBucket) Delete(id string) {
	log.Println("TCPConnBucket Delete id =", id)

	self.mu.Lock()
	delete(self.m, id)
	self.mu.Unlock()
}
func (self *TCPConnBucket) GetAll() map[string]*TCPConn {
	self.mu.RLock()
	defer self.mu.RUnlock()
	m := make(map[string]*TCPConn, len(self.m))
	for k, v := range self.m {
		m[k] = v
	}
	return m
}

func (self *TCPConnBucket) removeClosedTCPConn() {
	removeKey := make(map[string]struct{})
	for key, conn := range self.GetAll() {
		if conn.IsClosed() {
			removeKey[key] = struct{}{}
		}
	}
	for key := range removeKey {
		self.Delete(key)
	}
}
