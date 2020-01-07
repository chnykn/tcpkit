package main

import (
	"log"
	"tcp"
)


type CPLCallback struct{}


func (self *CPLCallback) OnConnected(conn *tcp.TCPConn) {
	log.Println("new conn: ", conn.GetRemoteIPAddress())
}

func (self *CPLCallback) OnDisconnected(conn *tcp.TCPConn) {
	log.Printf("%s disconnected \n", conn.GetRemoteIPAddress())
}

func (self *CPLCallback) OnError(conn *tcp.TCPConn, err error) {
	log.Println(err)
}

func (self *CPLCallback) OnMessage(conn *tcp.TCPConn, packet tcp.Packet) {
	//log.Println("receive: %s", string(packet.Bytes()))

	log.Printf("receive server packet: %v \n", tcp.PacketToString(packet))
	//TODO: ... 根据客户端的包，进行相应的操作，并返回相应的包 ...
	//conn.AsyncWritePacket(np)
}