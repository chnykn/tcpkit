package tcpkit

// CallBack interface used for various event handling during a connectio
type CallBack interface {

	//OnConnected when connection established
	OnConnected(conn *Conn)

	//OnMessage when message processing
	OnMessage(conn *Conn, p Packet)

	//OnDisconnected when connection disconnected
	OnDisconnected(conn *Conn)

	//OnError when error occurred.
	OnError(conn *Conn, err error)
}
