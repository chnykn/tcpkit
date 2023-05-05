package tcpkit

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

type BytePacket struct {
	id   []byte //长度6
	cmd  []byte //长度2
	size []byte //长度2
	body []byte //长度n
	sum  []byte //长度1
}

func NewBytePacket(id string, cmd []byte, body []byte) (*BytePacket, error) {
	if len(id) != 12 || len(cmd) != 2 {
		return nil, errors.New("packet's args invalid")
	}

	btsId, _ := hex.DecodeString(id)
	btsSize := make([]byte, 2)

	binary.LittleEndian.PutUint16(btsSize, uint16(len(body)))

	var buf bytes.Buffer
	buf.Write(btsId)
	buf.Write(cmd)
	buf.Write(btsSize)
	buf.Write(body)
	sum := CheckSum8(buf.Bytes(), 0, -1)

	return &BytePacket{
		id:   btsId,
		cmd:  cmd,
		size: btsSize,
		body: body,
		sum:  []byte{sum},
	}, nil
}

//----------------------------------------

func (o *BytePacket) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteByte('(')

	buf.Write(o.id)
	buf.Write(o.cmd)
	buf.Write(o.size)
	buf.Write(o.body)

	bts := buf.Bytes()
	sum := CheckSum8(bts, 1, len(bts))
	buf.WriteByte(sum)

	buf.WriteByte(')')

	return buf.Bytes()
}

func (o *BytePacket) ToString() string {
	var buf bytes.Buffer

	buf.WriteString("( Id:")
	buf.Write(o.id)

	buf.WriteString(", Cmd:")
	buf.Write(o.cmd)

	buf.WriteString(", Size:")
	buf.Write(o.size)

	buf.WriteString(", Body:")
	buf.Write(o.body)

	buf.WriteString(", Sum:")
	buf.Write(o.sum)

	buf.WriteString(" )")
	return buf.String()
}

//=============================================================================

type ByteProtocol struct {
}

func parsePacket(data []byte) (*BytePacket, error) {

	l := len(data)

	if l < 13 {
		return nil, errors.New("packet's length invalid")
	}

	//           '('                ')'
	if data[0] != 28 || data[l-1] != 29 {
		return nil, errors.New("packet's format invalid")
	}

	n := 1
	var id = data[n : n+12] //长度12

	n += 12
	var cmd = data[n : n+2] //长度2

	n += 2
	var btsSize = data[n : n+2] //长度2

	n += 2
	size := binary.LittleEndian.Uint16(btsSize)
	var body = data[n : n+int(size)] //长度size

	n += int(size)
	sum := data[n : n+1]
	sum1 := sum[0]
	sum2 := CheckSum8(data[1:n], 0, 16+int(size))
	if sum1 != sum2 {
		return nil, errors.New("packet's sum invalid")
	}

	return &BytePacket{
		id:   id,
		cmd:  cmd,
		size: btsSize,
		body: body,
		sum:  sum,
	}, nil
}

func (o *ByteProtocol) ReadPacket(reader io.Reader) (Packet, error) {
	rd := bufio.NewReader(reader)
	data, err := rd.ReadBytes('\n') // 读取的data末尾是'\n'
	if err != nil {
		return nil, err
	}

	return parsePacket(data)
}

func (o *ByteProtocol) WritePacket(writer io.Writer, packet Packet) error {
	_, err := writer.Write(packet.Bytes())
	return err
}
