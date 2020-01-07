package main

import (
	"log"
	"time"

	"github.com/chnykn/tcp"
)

func main() {

	srv := tcp.NewTCPServer("0.0.0.0:9001", &CPLCallback{}, &tcp.CharProtocol{})
	srv.SetReadDeadline(time.Second * 3600)

	log.Println("start listen...")
	log.Println(srv.ListenAndServe())

	select {}
}
