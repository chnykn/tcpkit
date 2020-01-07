package tcp

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

func byteToStr(b byte) string {
	res := strconv.FormatInt(int64(b), 16)
	if (len(res) < 2) {
		res = "0" + res
	}
	return res
}

//GetCRC
func GetCRC(data []byte) []byte {
	var c uint16
	var crc uint16 = 0xFFFF

	for i := 0; i < len(data); i++ {
		c = uint16(data[i]) & 0x00FF
		crc ^= c

		for j := 0; j < 8; j++ {
			if (crc & 0x0001) > 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, crc)

	var bts = buf.Bytes()
	if len(bts) == 2 {
		var res bytes.Buffer
		res.WriteString(byteToStr(bts[0]))
		res.WriteString(byteToStr(bts[1]))
		//fmt.Printf("字符串: %s  高位:%d  低位:%d  CRC:%s \n",  string(data), bts[0], bts[1], string(res.Bytes()))
		return res.Bytes()
	}

	return nil
}

