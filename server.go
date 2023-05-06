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

// NewServer Return an instance of Server
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

// ListenAndServe Create a TCPListener using the Server's tcpAddr, and call the Serve() method to start listening
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

// Serve -- Start listening using the specified TCPListener
func (o *Server) Serve(lsn *net.TCPListener) error {
	o.listener = lsn
	defer func() {
		if r := recover(); r != nil {
			log.Println("serve error", r)
		}
		_ = o.listener.Close()
	}()

	//Clean up invalid connections
	go func() {
		for {
			o.removeClosedConn()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	var tempDelay time.Duration

	for {
		select {
		case <-o.exitChan:
			return errors.New("server closed")
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
		tcpConn := o.newConn(conn, o.callback, o.protocol)
		tcpConn.setReadDeadline(o.readDeadline)
		tcpConn.setWriteDeadline(o.writeDeadline)
		o.connBucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (o *Server) removeClosedConn() {
	select {
	case <-o.exitChan:
		return
	default:
		o.connBucket.removeClosedConn()
	}
}

func (o *Server) newConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *Conn {
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

// Connect to other TCPServer using the specified callback and protocol
func (o *Server) Connect(ip string, callback CallBack, protocol Protocol) (*Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	tcpConn := o.newConn(conn, callback, protocol)
	return tcpConn, nil

}

// Close all connections, then close the TCPServer.
func (o *Server) Close() {
	defer o.listener.Close()
	for _, c := range o.connBucket.GetAll() {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

//----------------------------------------------------------------------------

func (o *Server) GetConn(key string) *Conn {
	return o.connBucket.Get(key)
}

func (o *Server) GetAllConn() []*Conn {
	result := []*Conn{}
	allConn := o.connBucket.GetAll()
	for _, conn := range allConn {
		result = append(result, conn)
	}
	return result
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
