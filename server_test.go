package tcpkit

import (
	"log"
	"testing"
)

type TestCallback struct{}

func (o *TestCallback) OnConnected(conn *Conn) {
	log.Println("new conn: ", conn.GetRemoteIPAddress())
}

func (o *TestCallback) OnDisconnected(conn *Conn) {
	log.Printf("%s disconnected \n", conn.GetRemoteIPAddress())
}

func (o *TestCallback) OnError(conn *Conn, err error) {
	log.Println(err)
}

func (o *TestCallback) OnMessage(conn *Conn, packet Packet) {
	//log.Println("receive: %s", string(packet.Bytes()))

	log.Printf("receive client packet: %v \n", packet.ToString())

	//TODO: ... 根据客户端的包，进行相应的操作，并返回相应的包 ...

	log.Println("server send: ", "0")

	pkt, err := NewBytePacket("123456789abc", []byte{0x00, 0x01}, []byte{0x00, 0x01})
	if err != nil {
		log.Println("pkt err: ", err.Error())
		return
	}

	_ = conn.AsyncWritePacket(pkt)
}

//==============================

func TestServer(t *testing.T) {

	srv := NewServer("0.0.0.0:9001", &TestCallback{}, &ByteProtocol{})
	//srv.SetReadDeadline(time.Second * 3600)

	log.Println("start listen...")
	log.Println(srv.ListenAndServe())

	select {}
}
