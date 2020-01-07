package tcp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	ErrCRCInvalid   = errors.New("packet's crc invalid")
	ErrCRCNotExists = errors.New("packet's crc not exist")
)

type CharPacket struct {
	typ  byte   //长度1
	num  []byte //长度4
	cmd  byte   //长度1
	body []byte //长度n
	crc  []byte //长度4
}

func NewCharPacket(typ byte, num []byte, cmd byte, body []byte) *CharPacket {
	if len(num) != 4 {
		panic(fmt.Errorf("num's len != 4"))
	}

	return &CharPacket{
		typ:  typ,
		num:  num,
		cmd:  cmd,
		body: body,
	}
}

func NewCharPacketEx(fromPacket Packet, body []byte) *CharPacket {
	return &CharPacket{
		typ:  fromPacket.Typ(),
		num:  fromPacket.Num(),
		cmd:  fromPacket.Cmd(),
		body: body,
	}
}

//---------------------------------------------------

func (self *CharPacket) Typ() byte {
	return self.typ
}

func (self *CharPacket) Num() []byte {
	return self.num
}

func (self *CharPacket) Cmd() byte {
	return self.cmd
}

func (self *CharPacket) Body() []byte {
	return self.body
}

func (self *CharPacket) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteByte(self.typ)
	buf.Write(self.num)
	buf.WriteByte(self.cmd)
	buf.Write(self.body)

	var crc = GetCRC(buf.Bytes())
	buf.WriteByte('#')
	buf.Write(crc)
	buf.WriteByte('\n')

	return buf.Bytes()
}

//=======================================================================

type CharProtocol struct {
}

func parsePacket(data []byte) (*CharPacket, error) {
	var n = -1
	for i := 0; i < len(data); i++ {
		if data[i] == '#' {
			n = i
			break
		}
	}
	if n < 0 {
		return nil, ErrCRCNotExists
	}

	var typ = data[0]    //长度1
	var num = data[1:5]  //长度4
	var cmd = data[5]    //长度1
	var body = data[6:n] //长度n-6

	//检查crc
	var crc1 = data[n+1:]
	if len(crc1) > 4 { // 末尾的写入'\n'即chr(10), 只有前两位是crc，第三位是chr(10)
		crc1 = crc1[:4]
	}
	var crc2 = GetCRC(data[0:n])
	if string(crc1) != string(crc2) {
		//log.Printf("CRC check error: %v , %v \n", crc1, crc2)
		return nil, ErrCRCInvalid
	}

	return NewCharPacket(typ, num, cmd, body), nil
}

func (self *CharProtocol) ReadPacket(reader io.Reader) (Packet, error) {
	rd := bufio.NewReader(reader)
	data, err := rd.ReadBytes('\n') // 读取的data末尾是'\n'
	if err != nil {
		return nil, err
	}

	return parsePacket(data)
}

func (self *CharProtocol) WritePacket(writer io.Writer, packet Packet) error {
	_, err := writer.Write(packet.Bytes())
	return err
}
