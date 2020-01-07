package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/chnykn/tcp"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "47.106.183.125:9001") //  127.0.0.1
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}
	tc := tcp.NewTCPConn(conn, &CPLCallback{}, &tcp.CharProtocol{})
	log.Println(tc.Serve())

	i := 0
	for {
		if tc.IsClosed() {
			break
		}

		msg := fmt.Sprintf("hello %d", i)
		log.Println("client send: ", msg)

		_ = tc.AsyncWritePacket(tcp.NewCharPacket(48, tcp.RandBytes(4), 97, []byte(msg)))

		i++
		time.Sleep(time.Second * 2)
	}
}
