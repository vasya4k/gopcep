package main

import (
	"fmt"
	"gopcep/pcep"
	"log"
	"net"
)

const msgTooShortErr = "recived msg is too short to parse common header and common object header out of it"

func handleTCPConn(conn net.Conn) {
	var s pcep.Session
	log.Println("Got connection", conn.RemoteAddr())
	defer conn.Close()
	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		if l < 4 {
			continue
		}

		s.HandleNewMsg(buff[:l], conn)
	}
}

func main() {
	// listen on all interfaces
	ln, err := net.Listen("tcp", "10.0.0.1:4189")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		// conn.Write([]byte(newmessage + "\n"))
		go handleTCPConn(conn)
	}
}
