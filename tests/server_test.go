package tests

import (
	"log"
	"testing"

	"github.com/chnykn/tcpkit"
)

func TestServer(t *testing.T) {

	srv := tcpkit.NewServer("0.0.0.0:9001", &ServerCallback{}, &tcpkit.ByteProtocol{})
	//srv.SetReadDeadline(time.Second * 3600)

	log.Println("start listen...")
	log.Println(srv.ListenAndServe())

	select {}
}
