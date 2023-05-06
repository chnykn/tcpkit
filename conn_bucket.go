package tcpkit

import (
	"sync"
)

// ConnBucket Used to store and manage Conn connections
type ConnBucket struct {
	m  map[string]*Conn
	mu *sync.RWMutex
}

func NewConnBucket() *ConnBucket {
	tcb := &ConnBucket{
		m:  make(map[string]*Conn),
		mu: new(sync.RWMutex),
	}
	return tcb
}

func (o *ConnBucket) Put(key string, c *Conn) {
	//log.Println("ConnBucket PUT key =", key)

	o.mu.Lock()
	if conn, ok := o.m[key]; ok {
		conn.Close()
	}
	o.m[key] = c
	o.mu.Unlock()
}

func (o *ConnBucket) Get(key string) *Conn {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if conn, ok := o.m[key]; ok {
		return conn
	}
	return nil
}

func (o *ConnBucket) Delete(key string) {
	//log.Println("ConnBucket Delete key =", key)

	o.mu.Lock()
	delete(o.m, key)
	o.mu.Unlock()
}

func (o *ConnBucket) GetAll() map[string]*Conn {
	o.mu.RLock()
	defer o.mu.RUnlock()
	m := make(map[string]*Conn, len(o.m))
	for k, v := range o.m {
		m[k] = v
	}
	return m
}

func (o *ConnBucket) removeClosedConn() {
	removeKey := make(map[string]struct{})
	for key, conn := range o.GetAll() {
		if conn.IsClosed() {
			removeKey[key] = struct{}{}
		}
	}
	for key := range removeKey {
		o.Delete(key)
	}
}
