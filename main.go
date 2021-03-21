package main

import (
	"gopcep/controller"
	"gopcep/grpcapi"
	"gopcep/pcep"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	controller := controller.Start()

	grpcapi.Start(&grpcapi.Config{
		Address: "127.0.0.1",
		Port:    "12345",
	}, controller)

	pcep.ListenForNewSession(controller)
}
