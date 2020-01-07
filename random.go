package tcp

import (
	"math/rand"
	"strconv"
	"time"
)

func RandomNum() []byte {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	result := strconv.FormatInt(int64(rnd.Intn(1000)), 10)

	for {
		if len(result) >= 3 {
			break
		}
		result = "0" + result
	}

	return []byte(result)
}