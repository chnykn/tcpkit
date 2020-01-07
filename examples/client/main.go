package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"tcp"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9001") //  127.0.0.1
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

		_ = tc.AsyncWritePacket(tcp.NewCharPacket(48, tcp.RandomNum(), 97, []byte(msg)))

		i++
		time.Sleep(time.Second*5)
	}
}

