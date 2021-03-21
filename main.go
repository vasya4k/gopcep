package main

import (
	"gopcep/controller"
	"gopcep/grpcapi"
	"gopcep/pcep"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
	})

	controller := controller.Start()

	grpcapi.Start(&grpcapi.Config{
		Address: "127.0.0.1",
		Port:    "12345",
	}, controller)

	pcep.ListenForNewSession(controller)
}
