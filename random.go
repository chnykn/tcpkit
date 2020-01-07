package tcp

import (
	"math/rand"
	"time"
)

// RandBytes 生成随机字符串
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
