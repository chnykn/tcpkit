package tcpkit

import (
	"math/rand"
	"time"
)

func CheckSum8(bts []byte, start, end int) byte {
	var res byte = 0
	if end < 0 {
		end = len(bts)
	}

	for i := start; i < end; i++ {
		res += bts[i]
	}

	return res
}

// Mpeg2Crc is an algorithm for CRC-32/MPEG-2 calculation
func Mpeg2Crc(data []byte) uint32 {
	crc := uint32(0xffffffff)

	for _, v := range data {
		crc ^= uint32(v) << 24
		for i := 0; i < 8; i++ {
			if (crc & 0x80000000) != 0 {
				crc = (crc << 1) ^ 0x04C11DB7
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// RandBytes Generate random string
func RandBytes(len int) []byte {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)

	var i int
	for {
		b := 36 + rnd.Intn(90) // 从36开始 绕开#(35)
		bytes[i] = byte(b)
		i += 1
		if i >= len {
			break
		}
	}

	return bytes
}
