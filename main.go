package main

import (
	"gopcep/pcep"
	"log"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

var routineCount int

const msgTooShortErr = "recived msg is too short to be able to parse common header "

func handleTCPConn(conn net.Conn) {
	defer conn.Close()
	routineCount++
	s := &pcep.Session{
		Conn:   conn,
		StopKA: make(chan struct{}),
	}
	s.Configure()
	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"remote_addr": conn.RemoteAddr().String(),
				"count":       routineCount,
			}).Error("connection read err")
			close(s.StopKA)
			return
		}
		if l < 4 {
			continue
		}
		s.HandleNewMsg(buff[:l])
	}
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
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
		if strings.Split(conn.RemoteAddr().String(), ":")[0] == "10.0.0.10" {
			logrus.WithFields(logrus.Fields{
				"remote_addr": conn.RemoteAddr().String(),
			}).Info("new connection")
			go handleTCPConn(conn)
		}
	}
}
