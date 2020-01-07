package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)



type BinPacket struct {
	typ byte     //长度1
	num []byte   //长度3
	cmd byte     //长度1
	body []byte  //长度n
}

func NewBinPacket(typ byte,  num []byte, cmd byte, body []byte) *BinPacket {
	return &BinPacket{
		typ: typ,
		num: num,
		cmd: cmd,
		body: body,
	}
}

//---------------------------------------------------

func (self *BinPacket) Typ() byte {
	return self.typ
}

func (self *BinPacket) Num() []byte {
	return self.num
}

func (self *BinPacket) Cmd() byte {
	return self.cmd
}

func (self *BinPacket) Body() []byte {
	return self.body
}


func (self *BinPacket) Bytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, self.typ)
	_ = binary.Write(&buf, binary.BigEndian, self.num)
	_ = binary.Write(&buf, binary.BigEndian, self.cmd)
	_ = binary.Write(&buf, binary.BigEndian, self.body)
	return buf.Bytes()
}


//=======================================================================

const headLength = 4

type BinProtocol struct {
	maxPacketSize uint32
}

func (p *BinProtocol) SetMaxPacketSize(n uint32) {
	p.maxPacketSize = n
}

func (self *BinProtocol) ReadPacket(reader io.Reader) (Packet, error) {
	return self.ReadPacketLimit(reader, self.maxPacketSize)
}

func (self *BinProtocol) ReadPacketLimit(reader io.Reader, size uint32) (Packet, error) {
	head := make([]byte, headLength)

	_, err := io.ReadFull(reader, head)
	if err != nil {
		return nil, err
	}

	packetLength := binary.BigEndian.Uint32(head)
	if size != 0 && packetLength > size {
		return nil, fmt.Errorf("packet too large(%v > %v)", packetLength, size)
	}

	buf := make([]byte, packetLength)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return NewBinPacket(buf[0], buf[1:4], buf[4], buf[5:]), nil
}

func (self *BinProtocol) WritePacket(writer io.Writer, packet Packet) error {
	var buf bytes.Buffer
	head := make([]byte, 4)
	data := packet.Bytes()

	binary.BigEndian.PutUint32(head, uint32(len(data)))
	_ = binary.Write(&buf, binary.BigEndian, head)
	_ = binary.Write(&buf, binary.BigEndian, data)

	_, err := writer.Write(buf.Bytes())
	return err
}


