package tcpkit

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

type Conn struct {
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
	extraData map[string]any
}

func NewConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *Conn {
	//_ = conn.SetKeepAlive(true)
	//_ = conn.SetKeepAlivePeriod(time.Second*30)

	result := &Conn{
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

func (o *Conn) Serve() error {
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("tcp conn(%v) Serve error, %v \n", o.GetRemoteIPAddress(), r)
		}
	}()

	if o.callback == nil || o.protocol == nil {
		err := fmt.Errorf("callback and protocol are not allowed to be nil")
		o.Close()
		return err
	}

	atomic.StoreInt32(&o.exitFlag, 1)
	o.callback.OnConnected(o)

	go o.readLoop()
	go o.writeLoop()
	go o.handleLoop()

	return nil
}

func (o *Conn) readLoop() {
	defer func() {
		recover()
		o.Close()
	}()

	for {
		select {
		case <-o.exitChan:
			return
		default:
			if o.readDeadline > 0 {
				_ = o.conn.SetReadDeadline(time.Now().Add(o.readDeadline))
			}
			p, err := o.protocol.ReadPacket(o.conn)
			if err != nil {
				if err != io.EOF {
					o.callback.OnError(o, err)
				}
				//If an error occurs in ReadPacket and the function exits, the current connection will be closed.
				//Close is present in the defer statement at the beginning of this function
				return
			}
			o.readChan <- p
		}
	}
}

func (o *Conn) writeLoop() {
	defer func() {
		recover()
		o.Close()
	}()

	for pkt := range o.writeChan {
		if pkt == nil {
			continue
		}
		if o.writeDeadline > 0 {
			_ = o.conn.SetWriteDeadline(time.Now().Add(o.writeDeadline))
		}
		if err := o.protocol.WritePacket(o.conn, pkt); err != nil {
			o.callback.OnError(o, err)
			return
		}
	}
}

func (o *Conn) handleLoop() {
	defer func() {
		recover()
		o.Close()
	}()

	for p := range o.readChan {
		if p == nil {
			continue
		}
		o.callback.OnMessage(o, p)
	}
}

//----------------------------------------------------------------------------

func (o *Conn) ReadPacket() (Packet, error) {
	if o.protocol == nil {
		return nil, errors.New("no protocol impl")
	}
	return o.protocol.ReadPacket(o.conn)
}

func (o *Conn) AsyncWritePacket(p Packet) error {
	if o.IsClosed() {
		return ErrConnClosing
	}
	select {
	case o.writeChan <- p:
		return nil
	default:
		return ErrBufferFull
	}
}

func (o *Conn) AsyncWritePacketWithTimeout(p Packet, sec int) error {
	if o.IsClosed() {
		return ErrConnClosing
	}
	select {
	case o.writeChan <- p:
		return nil
	case <-time.After(time.Second * time.Duration(sec)):
		return ErrWriteTimeout
	}
}

//----------------------------------------------------------------------------

func (o *Conn) Close() {
	o.closeOnce.Do(func() {
		atomic.StoreInt32(&o.exitFlag, 0)
		close(o.exitChan)
		close(o.writeChan)
		close(o.readChan)
		if o.callback != nil {
			o.callback.OnDisconnected(o)
		}
		_ = o.conn.Close()
	})
}

func (o *Conn) IsClosed() bool {
	return atomic.LoadInt32(&o.exitFlag) == 0
}

func (o *Conn) GetRawConn() *net.TCPConn {
	return o.conn
}

//----------------------------------------------------------------------------

func (o *Conn) GetLocalAddr() net.Addr {
	return o.conn.LocalAddr()
}

// GetLocalIPAddress get the local IP address of a socket connection.
func (o *Conn) GetLocalIPAddress() string {
	return strings.Split(o.GetLocalAddr().String(), ":")[0]
}

func (o *Conn) GetRemoteAddr() net.Addr {
	return o.conn.RemoteAddr()
}

func (o *Conn) GetRemoteIPAddress() string {
	return strings.Split(o.GetRemoteAddr().String(), ":")[0]
}

//----------------------------------------------------------------------------

func (o *Conn) setReadDeadline(t time.Duration) {
	o.readDeadline = t
}

func (o *Conn) setWriteDeadline(t time.Duration) {
	o.writeDeadline = t
}

//----------------------------------------------------------------------------

func (o *Conn) SetExtraData(key string, data any) {
	if o.extraData == nil {
		o.extraData = make(map[string]any)
	}
	o.extraData[key] = data
}

func (o *Conn) GetExtraData(key string) any {
	if o.extraData != nil {
		if data, ok := o.extraData[key]; ok {
			return data
		}
	}
	return nil
}

func (o *Conn) ExtraData() map[string]any {
	return o.extraData
}
