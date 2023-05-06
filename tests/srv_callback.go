package tests

import (
	"log"

	"github.com/chnykn/tcpkit"
)

type ServerCallback struct{}

func (o *ServerCallback) OnConnected(conn *tcpkit.Conn) {
	log.Println("new conn: ", conn.GetRemoteIPAddress())
}

func (o *ServerCallback) OnDisconnected(conn *tcpkit.Conn) {
	log.Printf("%s disconnected \n", conn.GetRemoteIPAddress())
}

func (o *ServerCallback) OnError(conn *tcpkit.Conn, err error) {
	log.Println(err)
}

func (o *ServerCallback) OnMessage(conn *tcpkit.Conn, packet tcpkit.Packet) {
	//log.Println("receive: %s", string(packet.Bytes()))

	log.Printf("receive client packet: %v \n", packet.ToString())

	//TODO: ... 根据客户端的包，进行相应的操作，并返回相应的包 ...

	log.Println("server send: ", "0")

	pkt, err := tcpkit.NewBytePacket("123456789abc", []byte{0x00, 0x01}, []byte{0x00, 0x01})
	if err != nil {
		log.Println("pkt err: ", err.Error())
		return
	}

	_ = conn.AsyncWritePacket(pkt)
}
