package tcpkit

import (
	"errors"
	"log"
	"net"
	"os"
	"time"
)

const (
	//tcp conn max packet size
	defaultMaxPacketSize = 1024 << 10 //1MB

	readChanSize  = 100
	writeChanSize = 100
)

var (
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", log.Lshortfile)
}

// Server 结构定义
type Server struct {
	//TCP address to listen on
	tcpAddr string

	//the listener
	listener *net.TCPListener

	//callback is an interface
	//it's used to process the connect establish, close and data receive
	callback CallBack
	protocol Protocol

	//if o is shutdown, close the channel used to inform all session to exit.
	exitChan chan struct{}

	readDeadline  time.Duration
	writeDeadline time.Duration
	connBucket    *ConnBucket
}

// NewServer 返回一个TCPServer实例
func NewServer(tcpAddr string, callback CallBack, protocol Protocol) *Server {
	return &Server{
		tcpAddr:  tcpAddr,
		callback: callback,
		protocol: protocol,

		connBucket: NewConnBucket(),
		exitChan:   make(chan struct{}),
	}
}

//======================================================================================

// ListenAndServe 使用TCPServer的tcpAddr创建TCPListener并调用Server()方法开启监听
func (o *Server) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", o.tcpAddr)
	if err != nil {
		return err
	}

	lsn, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}

	go o.Serve(lsn)
	return nil
}

// Serve 使用指定的TCPListener开启监听
func (o *Server) Serve(lsn *net.TCPListener) error {
	o.listener = lsn
	defer func() {
		if r := recover(); r != nil {
			log.Println("Serve error", r)
		}
		_ = o.listener.Close()
	}()

	//清理无效连接
	go func() {
		for {
			o.removeClosedTCPConn()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	var tempDelay time.Duration

	for {
		select {
		case <-o.exitChan:
			return errors.New("Server Closed")
		default:
		}

		conn, err := o.listener.AcceptTCP()
		if err != nil {
			err, ok := err.(net.Error)
			if ok && err.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}

			log.Println("listener error:", err.Error())
			return err
		}

		tempDelay = 0
		tcpConn := o.newTCPConn(conn, o.callback, o.protocol)
		tcpConn.setReadDeadline(o.readDeadline)
		tcpConn.setWriteDeadline(o.writeDeadline)
		o.connBucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (o *Server) removeClosedTCPConn() {
	select {
	case <-o.exitChan:
		return
	default:
		o.connBucket.removeClosedTCPConn()
	}
}

func (o *Server) newTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *Conn {
	if callback == nil {
		// if the handler is nil, use o handler
		callback = o.callback
	}

	if protocol == nil {
		protocol = o.protocol
	}

	c := NewConn(conn, callback, protocol)
	_ = c.Serve()

	return c
}

//----------------------------------------------------------------------------

// Connect 使用指定的callback和protocol连接其他TCPServer，返回TCPConn
func (o *Server) Connect(ip string, callback CallBack, protocol Protocol) (*Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	tcpConn := o.newTCPConn(conn, callback, protocol)
	return tcpConn, nil

}

// Close 首先关闭所有连接，然后关闭TCPServer
func (o *Server) Close() {
	defer o.listener.Close()
	for _, c := range o.connBucket.GetAll() {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

//----------------------------------------------------------------------------

func (o *Server) GetAllTCPConn() []*Conn {
	result := []*Conn{}
	allConn := o.connBucket.GetAll()
	for _, conn := range allConn {
		result = append(result, conn)
	}
	return result
}

func (o *Server) GetTCPConn(key string) *Conn {
	return o.connBucket.Get(key)
}

//----------------------------------------------------------------------------

func (o *Server) SetReadDeadline(t time.Duration) {
	o.readDeadline = t
}

func (o *Server) SetWriteDeadline(t time.Duration) {
	o.writeDeadline = t
}

func (o *Server) GetAddr() string {
	return o.tcpAddr
}
