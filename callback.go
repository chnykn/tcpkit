package tcpkit

// CallBack 是一个回调接口，用于连接的各种事件处理
type CallBack interface {

	//OnConnected 链接建立回调
	OnConnected(conn *Conn)

	//OnMessage 消息处理回调
	OnMessage(conn *Conn, p Packet)

	//OnDisconnected 链接断开回调
	OnDisconnected(conn *Conn)

	//OnError 错误回调
	OnError(conn *Conn, err error)
}
