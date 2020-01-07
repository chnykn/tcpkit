package tcp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrConnClosing  = errors.New("use of closed network connection")
	ErrBufferFull   = errors.New("the async send buffer is full")
	ErrWriteTimeout = errors.New("async write packet timeout")
)

type TCPConn struct {
	callback CallBack
	protocol Protocol

	conn      *net.TCPConn
	readChan  chan Packet
	writeChan chan Packet

	readDeadline  time.Duration
	writeDeadline time.Duration

	exitChan  chan struct{}
	closeOnce sync.Once
	exitFlag  int32
	extraData map[string]interface{}
}

func NewTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConn {
	//_ = conn.SetKeepAlive(true)
	//_ = conn.SetKeepAlivePeriod(time.Second*30)

	result := &TCPConn{
		conn:     conn,
		callback: callback,
		protocol: protocol,

		readChan:  make(chan Packet, readChanSize),
		writeChan: make(chan Packet, writeChanSize),

		exitChan: make(chan struct{}),
		exitFlag: 0,
	}
	return result
}

//======================================================================================

func (self *TCPConn) Serve() error {
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("tcp conn(%v) Serve error, %v \n", self.GetRemoteIPAddress(), r)
		}
	}()

	if self.callback == nil || self.protocol == nil {
		err := fmt.Errorf("callback and protocol are not allowed to be nil")
		self.Close()
		return err
	}

	atomic.StoreInt32(&self.exitFlag, 1)
	self.callback.OnConnected(self)

	go self.readLoop()
	go self.writeLoop()
	go self.handleLoop()

	return nil
}

func (self *TCPConn) readLoop() {
	defer func() {
		recover()
		self.Close()
	}()

	for {
		select {
			case <-self.exitChan:
				return
			default:
				if self.readDeadline > 0 {
					_ = self.conn.SetReadDeadline(time.Now().Add(self.readDeadline))
				}
				p, err := self.protocol.ReadPacket(self.conn)
				if err != nil {
					if err != io.EOF {
						self.callback.OnError(self, err)
					}
					return
				}
				self.readChan <- p
		}
	}
}

func (self *TCPConn) writeLoop() {
	defer func() {
		recover()
		self.Close()
	}()

	for pkt := range self.writeChan {
		if pkt == nil {
			continue
		}
		if self.writeDeadline > 0 {
			_ = self.conn.SetWriteDeadline(time.Now().Add(self.writeDeadline))
		}
		if err := self.protocol.WritePacket(self.conn, pkt); err != nil {
			self.callback.OnError(self, err)
			return
		}
	}
}

func (self *TCPConn) handleLoop() {
	defer func() {
		recover()
		self.Close()
	}()

	for p := range self.readChan {
		if p == nil {
			continue
		}
		self.callback.OnMessage(self, p)
	}
}

//----------------------------------------------------------------------------

func (self *TCPConn) ReadPacket() (Packet, error) {
	if self.protocol == nil {
		return nil, errors.New("no protocol impl")
	}
	return self.protocol.ReadPacket(self.conn)
}


func (self *TCPConn) AsyncWritePacket(p Packet) error {
	if self.IsClosed() {
		return ErrConnClosing
	}
	select {
		case self.writeChan <- p:
			return nil
		default:
			return ErrBufferFull
	}
}

func (self *TCPConn) AsyncWritePacketWithTimeout(p Packet, sec int) error {
	if self.IsClosed() {
		return ErrConnClosing
	}
	select {
		case self.writeChan <- p:
			return nil
		case <-time.After(time.Second * time.Duration(sec)):
			return ErrWriteTimeout
	}
}

//----------------------------------------------------------------------------

func (self *TCPConn) Close() {
	self.closeOnce.Do(func() {
		atomic.StoreInt32(&self.exitFlag, 0)
		close(self.exitChan)
		close(self.writeChan)
		close(self.readChan)
		if self.callback != nil {
			self.callback.OnDisconnected(self)
		}
		_ = self.conn.Close()
	})
}

func (self *TCPConn) IsClosed() bool {
	return atomic.LoadInt32(&self.exitFlag) == 0
}

func (self *TCPConn) GetRawConn() *net.TCPConn {
	return self.conn
}

//----------------------------------------------------------------------------

func (self *TCPConn) GetLocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

//LocalIPAddress 返回socket连接本地的ip地址
func (self *TCPConn) GetLocalIPAddress() string {
	return strings.Split(self.GetLocalAddr().String(), ":")[0]
}

func (self *TCPConn) GetRemoteAddr() net.Addr {
	return self.conn.RemoteAddr()
}

func (self *TCPConn) GetRemoteIPAddress() string {
	return strings.Split(self.GetRemoteAddr().String(), ":")[0]
}

//----------------------------------------------------------------------------

func (self *TCPConn) setReadDeadline(t time.Duration) {
	self.readDeadline = t
}

func (self *TCPConn) setWriteDeadline(t time.Duration) {
	self.writeDeadline = t
}

//----------------------------------------------------------------------------

func (self *TCPConn) SetExtraData(key string, data interface{}) {
	if self.extraData == nil {
		self.extraData = make(map[string]interface{})
	}
	self.extraData[key] = data
}

func (self *TCPConn) GetExtraData(key string) interface{} {
	if data, ok := self.extraData[key]; ok {
		return data
	}
	return nil
}




