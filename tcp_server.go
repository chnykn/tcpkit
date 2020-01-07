package tcp

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

//TCPServer 结构定义
type TCPServer struct {
	//TCP address to listen on
	tcpAddr string

	//the listener
	listener *net.TCPListener

	//callback is an interface
	//it's used to process the connect establish, close and data receive
	callback CallBack
	protocol Protocol

	//if self is shutdown, close the channel used to inform all session to exit.
	exitChan chan struct{}

	readDeadline  time.Duration
	writeDeadline time.Duration
	connBucket    *TCPConnBucket
}

//NewTCPServer 返回一个TCPServer实例
func NewTCPServer(tcpAddr string, callback CallBack, protocol Protocol) *TCPServer {
	return &TCPServer{
		tcpAddr:  tcpAddr,
		callback: callback,
		protocol: protocol,

		connBucket: NewTCPConnBucket(),
		exitChan:   make(chan struct{}),
	}
}

//======================================================================================

//ListenAndServe 使用TCPServer的tcpAddr创建TCPListener并调用Server()方法开启监听
func (self *TCPServer) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.tcpAddr)
	if err != nil {
		return err
	}

	lsn, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}

	go self.Serve(lsn)
	return nil
}

//Serve 使用指定的TCPListener开启监听
func (self *TCPServer) Serve(lsn *net.TCPListener) error {
	self.listener = lsn
	defer func() {
		if r := recover(); r != nil {
			log.Println("Serve error", r)
		}
		_ = self.listener.Close()
	}()

	//清理无效连接
	go func() {
		for {
			self.removeClosedTCPConn()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	var tempDelay time.Duration

	for {
		select {
		case <-self.exitChan:
			return errors.New("TCPServer Closed")
		default:
		}

		conn, err := self.listener.AcceptTCP()
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
		tcpConn := self.newTCPConn(conn, self.callback, self.protocol)
		tcpConn.setReadDeadline(self.readDeadline)
		tcpConn.setWriteDeadline(self.writeDeadline)
		self.connBucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (self *TCPServer) removeClosedTCPConn() {
	select {
	case <-self.exitChan:
		return
	default:
		self.connBucket.removeClosedTCPConn()
	}
}

func (self *TCPServer) newTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConn {
	if callback == nil {
		// if the handler is nil, use self handler
		callback = self.callback
	}

	if protocol == nil {
		protocol = self.protocol
	}

	c := NewTCPConn(conn, callback, protocol)
	_ = c.Serve()

	return c
}

//----------------------------------------------------------------------------

//Connect 使用指定的callback和protocol连接其他TCPServer，返回TCPConn
func (self *TCPServer) Connect(ip string, callback CallBack, protocol Protocol) (*TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	tcpConn := self.newTCPConn(conn, callback, protocol)
	return tcpConn, nil

}

//Close 首先关闭所有连接，然后关闭TCPServer
func (self *TCPServer) Close() {
	defer self.listener.Close()
	for _, c := range self.connBucket.GetAll() {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

//----------------------------------------------------------------------------

func (self *TCPServer) GetAllTCPConn() []*TCPConn {
	result := []*TCPConn{}
	allConn := self.connBucket.GetAll()
	for _, conn := range allConn {
		result = append(result, conn)
	}
	return result
}

func (self *TCPServer) GetTCPConn(key string) *TCPConn {
	return self.connBucket.Get(key)
}

//----------------------------------------------------------------------------

func (self *TCPServer) SetReadDeadline(t time.Duration) {
	self.readDeadline = t
}

func (self *TCPServer) SetWriteDeadline(t time.Duration) {
	self.writeDeadline = t
}

func (self *TCPServer) GetAddr() string {
	return self.tcpAddr
}
