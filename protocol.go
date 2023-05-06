package tcpkit

import (
	"io"
)

type Packet interface {
	Bytes() []byte
	ToString() string
}

type Protocol interface {
	ReadPacket(reader io.Reader) (Packet, error)
	WritePacket(writer io.Writer, packet Packet) error
}
