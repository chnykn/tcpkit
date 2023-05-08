package tcpkit

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	idBtsLen = 6
	idStrLen = 12

	cmdBtsLen = 2
	cmdStrLen = 4

	sizeBtsLen = 2
	sumBtsLen  = 1

	// starting & delimiter (2 digits), ID (6 digits), command (2 digits), length of content (2 digits),
	// content (unknown, but at least 0), checksum (1 digit). Total of at least 13 digits."
	minPktBtsLen = 13
)

type BinPacket struct {
	id   []byte //size 6
	cmd  []byte //size 2
	size []byte //size 2
	cont []byte //size n //content
	sum  byte   //size 1
}

func NewBytePacket(id string, cmd string, content []byte) (*BinPacket, error) {
	if len(id) != idStrLen {
		return nil, fmt.Errorf("packet's id length must be %d", idStrLen)
	}

	if len(cmd) != cmdStrLen {
		return nil, fmt.Errorf("packet's cmd length must be %d", cmdStrLen)
	}

	var buf bytes.Buffer

	idBts, _ := hex.DecodeString(id)
	buf.Write(idBts)

	cmdBts, _ := hex.DecodeString(cmd)
	buf.Write(cmdBts)

	sizeBts := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBts, uint16(len(content)))
	buf.Write(sizeBts)

	buf.Write(content)
	sum := BCCVerify(buf.Bytes(), 0, -1)

	return &BinPacket{
		id:   idBts,
		cmd:  cmdBts,
		size: sizeBts,
		cont: content,
		sum:  sum,
	}, nil
}

//----------------------------------------

func (o *BinPacket) Id() string {
	return hex.EncodeToString(o.id)
}

func (o *BinPacket) Cmd() string {
	return hex.EncodeToString(o.cmd)
}

func (o *BinPacket) Size() int { //uint16
	size := binary.BigEndian.Uint16(o.size)
	return int(size)
}

func (o *BinPacket) Content() string { //uint16
	return string(o.cont)
}

func (o *BinPacket) ContentBytes() []byte { //uint16
	return o.cont
}

func (o *BinPacket) Checksum() int {
	return int(o.sum)
}

func (o *BinPacket) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteByte('(')

	buf.Write(o.id)
	buf.Write(o.cmd)
	buf.Write(o.size)
	buf.Write(o.cont)

	bts := buf.Bytes()
	sum := BCCVerify(bts, 1, len(bts))
	buf.WriteByte(sum)

	buf.WriteByte(')')

	return buf.Bytes()
}

func (o *BinPacket) ToString() string {
	var buf bytes.Buffer

	buf.WriteString("( Id:")
	buf.WriteString(o.Id())

	buf.WriteString(", Cmd:")
	buf.WriteString(o.Cmd())

	buf.WriteString(", Size:")
	buf.WriteString(strconv.Itoa(o.Size()))

	buf.WriteString(", Content:")
	buf.WriteString(o.Content())

	buf.WriteString(", Sum:")
	buf.WriteString(strconv.Itoa(o.Checksum()))

	buf.WriteString(" )")
	return buf.String()
}

//=============================================================================

type BinProtocol struct {
	LogEnabled bool
}

/*
check if there is 0x3d in the data, if so, then determine what data is followed by 0x3D.
If the data received is 0x3D 0x15, 0x3D 0x14, 0x3D 0x00,
convert these data to 0x28, 0x29, 0x3D respectively.
*/
func escapePacket(data []byte) []byte {
	res := data
	i := 0
	h := len(res) - 1
	for {
		if (res[i] == 0x3D) && (i < h) {
			escaped := false

			if res[i+1] == 0x00 {
				escaped = true
			} else if res[i+1] == 0x15 {
				res[i] = 0x28
				escaped = true
			} else if res[i+1] == 0x14 {
				res[i] = 0x29
				escaped = true
			}

			if escaped {
				res = append(res[:i+1], res[i+2:]...)
				h -= 1
			}
		}

		i += 1
		if i >= h {
			break
		}
	}

	return res
}

func parsePacket(data []byte, logEnabled bool) (*BinPacket, error) {

	data = escapePacket(data)

	l := len(data)

	if l < minPktBtsLen {
		return nil, fmt.Errorf("packet length cannot be less than %d", minPktBtsLen)
	}

	//           '('                  ')'
	if data[0] != 0x28 || data[l-1] != 0x29 {
		return nil, fmt.Errorf("packet must start with '(' and end with ')'")
	}

	n := 1
	var id = data[n : n+idBtsLen]

	n += idBtsLen
	var cmd = data[n : n+cmdBtsLen]

	pktName := hex.EncodeToString(cmd)
	if logEnabled {
		now := time.Now().Format("01-02 15.04.05.000")
		pktName = fmt.Sprintf("%s#%s.bin", pktName, now)
		filename := fmt.Sprintf("./log/%s", pktName)
		_ = os.WriteFile(filename, data, 0666)
	}

	n += cmdBtsLen
	var btsSize = data[n : n+sizeBtsLen]

	n += sizeBtsLen
	size := int(binary.BigEndian.Uint16(btsSize))
	if n+size > l-2 { // delimiter ) and sum, 2 digits
		return nil, fmt.Errorf("the size of the content in the packet is incorrect")
	}
	var content = data[n : n+size]

	n += size
	sum := data[n : n+1]
	sum1 := sum[0]
	sum2 := BCCVerify(data, 1, n)
	if sum1 != sum2 {
		//fmt.Printf("checksum is incorrect. Expected %x, but %x \n", sum1, sum2)
		return nil, fmt.Errorf("%s checksum is incorrect. Expected %x, but %x", pktName, sum1, sum2)
	}

	return &BinPacket{
		id:   id,
		cmd:  cmd,
		size: btsSize,
		cont: content,
		sum:  sum1,
	}, nil
}

func (o *BinProtocol) ReadPacket(reader io.Reader) (Packet, error) {
	rd := bufio.NewReader(reader)
	data, err := rd.ReadBytes(')') // end of the 'data' is a ')'
	if err != nil {
		return nil, err
	}

	return parsePacket(data, o.LogEnabled)
}

func (o *BinProtocol) WritePacket(writer io.Writer, packet Packet) error {
	_, err := writer.Write(packet.Bytes())
	return err
}
