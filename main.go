package main

import (
	"fmt"
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
		Conn:         conn,
		StopKA:       make(chan struct{}),
		Keepalive:    30,
		PLSPIDToName: make(map[uint32]string),
		LSPs:         make(map[string]*pcep.LSP),
	}
	// gAPI.StorePSessions(conn.RemoteAddr().String(), s)

	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"remote_addr": conn.RemoteAddr().String(),
				"count":       routineCount,
			}).Error(err)
			close(s.StopKA)
			return
		}
		if l < 4 {
			fmt.Printf("Something is not right %d", l)
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
	ln, err := net.Listen("tcp", "192.168.1.14:4189")
	if err != nil {
		log.Fatalln(err)
	}
	// gAPI := grpcapi.Start(&grpcapi.Config{Address: "127.0.0.1", Port: "12345"})
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
