package tcpkit

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
)

type BytePacket struct {
	id   []byte //size 6
	cmd  []byte //size 2
	size []byte //size 2
	cont []byte //size n //content
	sum  byte   //size 1
}

func NewBytePacket(id string, cmd []byte, content []byte) (*BytePacket, error) {
	if len(id) != 12 || len(cmd) != 2 {
		return nil, errors.New("packet's args invalid")
	}

	var buf bytes.Buffer

	btsId, _ := hex.DecodeString(id)
	buf.Write(btsId)

	buf.Write(cmd)

	btsSize := make([]byte, 2)
	binary.BigEndian.PutUint16(btsSize, uint16(len(content)))
	buf.Write(btsSize)

	buf.Write(content)
	sum := CheckSum8(buf.Bytes(), 0, -1)

	return &BytePacket{
		id:   btsId,
		cmd:  cmd,
		size: btsSize,
		cont: content,
		sum:  sum,
	}, nil
}

//----------------------------------------

func (o *BytePacket) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteByte('(')

	buf.Write(o.id)
	buf.Write(o.cmd)
	buf.Write(o.size)
	buf.Write(o.cont)

	bts := buf.Bytes()
	sum := CheckSum8(bts, 1, len(bts))
	buf.WriteByte(sum)

	buf.WriteByte(')')

	return buf.Bytes()
}

func (o *BytePacket) ToString() string {
	var buf bytes.Buffer

	buf.WriteString("( Id:")
	buf.WriteString(hex.EncodeToString(o.id))

	buf.WriteString(", Cmd:")
	buf.WriteString(hex.EncodeToString(o.cmd))

	buf.WriteString(", Size:")
	size := binary.BigEndian.Uint16(o.size)
	buf.WriteString(strconv.Itoa(int(size)))

	buf.WriteString(", Content:")
	buf.Write(o.cont)

	buf.WriteString(", Sum:")
	buf.WriteString(strconv.Itoa(int(o.sum)))

	buf.WriteString(" )")
	return buf.String()
}

//=============================================================================

type ByteProtocol struct {
}

func parsePacket(data []byte) (*BytePacket, error) {

	l := len(data)

	// starting & delimiter (2 digits), ID (6 digits), command (2 digits), length of content (2 digits),
	// content (unknown, but at least 0), checksum (1 digit). Total of at least 13 digits."
	if l < 13 {
		return nil, errors.New("packet's length invalid")
	}

	//           '('                  ')'
	if data[0] != 0x28 || data[l-1] != 0x29 {
		return nil, errors.New("packet's format invalid")
	}

	n := 1
	var id = data[n : n+6]

	n += 6
	var cmd = data[n : n+2]

	n += 2
	var btsSize = data[n : n+2]

	n += 2
	size := int(binary.BigEndian.Uint16(btsSize))
	if n+size > l-1 { //
		return nil, errors.New("packet's size invalid")
	}
	var content = data[n : n+size]

	n += size
	sum := data[n : n+1]
	sum1 := sum[0]
	sum2 := CheckSum8(data, 1, n)
	if sum1 != sum2 {
		return nil, errors.New("packet's sum invalid")
	}

	return &BytePacket{
		id:   id,
		cmd:  cmd,
		size: btsSize,
		cont: content,
		sum:  sum1,
	}, nil
}

func (o *ByteProtocol) ReadPacket(reader io.Reader) (Packet, error) {
	rd := bufio.NewReader(reader)
	data, err := rd.ReadBytes(')') // end of the 'data' is a ')'
	if err != nil {
		return nil, err
	}

	return parsePacket(data)
}

func (o *ByteProtocol) WritePacket(writer io.Writer, packet Packet) error {
	_, err := writer.Write(packet.Bytes())
	return err
}
