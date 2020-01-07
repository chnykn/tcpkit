package main

import (
	"log"
	"time"

	"tcp"
)

func main() {
	str := "12abc"
	tcp.GetCRC([]byte(str)) //EB6F  //235 111

	str = "xyzXYZ"
	tcp.GetCRC([]byte(str))

	str = "9876543"
	tcp.GetCRC([]byte(str))

	str = "!@#$%^&*():;',."
	tcp.GetCRC([]byte(str))


	srv := tcp.NewTCPServer("0.0.0.0:9001", &CPLCallback{}, &tcp.CharProtocol{})
	srv.SetReadDeadline(time.Second * 3600)

	log.Println("start listen...")
	log.Println(srv.ListenAndServe())

	select {}
}

