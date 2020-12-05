package main

import (
	"gopcep/grpcapi"
	"gopcep/pcep"
	"log"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

const msgTooShortErr = "recived msg is too short to be able to parse common header "

func startPCEPSession(conn net.Conn, gAPI *grpcapi.GRPCAPI) {
	session := pcep.NewSession(conn)
	gAPI.StorePSessions(conn.RemoteAddr().String(), session)

	defer func() {
		err := conn.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic":       "closing connection",
				"remote_addr": conn.RemoteAddr().String(),
			}).Error(err)
		}
		gAPI.DeletePSessions(conn.RemoteAddr().String())
	}()

	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"remote_addr": conn.RemoteAddr().String(),
			}).Error(err)
			close(session.StopKA)
			return
		}
		session.HandleNewMsg(buff[:l])
	}
}

func remoteIPFiltered(conn net.Conn, filteredIPs []string) bool {
	for _, ip := range filteredIPs {
		if strings.Split(conn.RemoteAddr().String(), ":")[0] == ip {
			return true
		}
	}
	return false
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

	gAPI := grpcapi.Start(&grpcapi.Config{Address: "127.0.0.1", Port: "12345"})

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		if remoteIPFiltered(conn, []string{"10.0.0.10"}) {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"remote_addr": conn.RemoteAddr().String(),
		}).Info("new connection")
		go startPCEPSession(conn, gAPI)
	}
}
