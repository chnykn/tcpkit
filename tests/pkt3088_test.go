package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/chnykn/tcpkit"
)

var pkt3088files = []string{
	"./log/3088#05-08 07.10.57.502.bin",
	"./log/3088#05-08 07.21.16.122.bin",
	"./log/3088#05-08 07.25.10.455.bin",
}

func TestPacket3088(t *testing.T) {
	var ptl = &tcpkit.ByteProtocol{}
	var pkt tcpkit.Packet

	for i := 0; i < len(pkt3088files); i++ {
		f, _ := os.Open(pkt3088files[i])
		if f == nil {
			continue
		}

		bts, err := io.ReadAll(f)
		if err == nil {
			buf := bytes.NewBuffer(bts)
			pkt, err = ptl.ReadPacket(buf)
		}
		if err == nil && pkt != nil {
			bp := pkt.(*tcpkit.BytePacket)
			trip, er := ParseTrip(bp.Id(), bp.ContentBytes())

			if er != nil {
				fmt.Println("ParseTrip ERR:", err.Error())
			} else {
				fmt.Printf("trip:%v\n", *trip)
			}
		}

		f.Close()
	}

}
