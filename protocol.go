package tcp

import (
	"bytes"
	"io"
)


type Packet interface {
	Typ() byte
	Num() []byte
	Cmd() byte
	Body() []byte

	Bytes() []byte
}


type Protocol interface {
	ReadPacket(reader io.Reader) (Packet, error)
	WritePacket(writer io.Writer, packet Packet) error
}


func PacketToString(p Packet) string  {
	var buf bytes.Buffer
	buf.WriteString("{ typ:")
	buf.WriteByte(p.Typ())
	buf.WriteString(", num:")
	buf.Write(p.Num())
	buf.WriteString(", cmd:")
	buf.WriteByte(p.Cmd())
	buf.WriteString(", body:")
	buf.Write(p.Body())
	buf.WriteString(" }")
	return buf.String()
}

