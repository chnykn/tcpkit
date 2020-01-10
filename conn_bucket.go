package tcp

import (
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

func (self *TCPConnBucket) Put(key string, c *TCPConn) {
	//log.Println("TCPConnBucket PUT key =", key)

	self.mu.Lock()
	if conn, ok := self.m[key]; ok {
		conn.Close()
	}
	self.m[key] = c
	self.mu.Unlock()
}

func (self *TCPConnBucket) Get(key string) *TCPConn {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if conn, ok := self.m[key]; ok {
		return conn
	}
	return nil
}

func (self *TCPConnBucket) Delete(key string) {
	//log.Println("TCPConnBucket Delete key =", key)

	self.mu.Lock()
	delete(self.m, key)
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
