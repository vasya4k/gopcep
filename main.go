package main

import (
	"gopcep/controller"
	"gopcep/grpcapi"
	"gopcep/pcep"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type cfg struct {
	grpcapi grpcapi.Config
	pcep    pcep.Cfg
}

func appCfg(cfgPath string) *cfg {
	viper.SetConfigName("gopcep")
	viper.AddConfigPath(cfgPath)
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic":    "config",
			"event":    "cfg read error",
			"cfg_path": cfgPath,
		}).Fatal(err)
	}

	return &cfg{
		pcep: pcep.Cfg{
			ListenAddr: viper.GetString("pcep.listen_addr"),
			ListenPort: viper.GetString("pcep.listen_port"),
			Keepalive:  uint8(viper.GetUint32("pcep.keepalive")),
			PCClients:  viper.GetStringSlice("pcep.pcc_clients"),
		},
		grpcapi: grpcapi.Config{
			ListenAddr: viper.GetString("grpcapi.listen_addr"),
			ListenPort: viper.GetString("grpcapi.listen_port"),
			Tokens:     viper.GetStringSlice("grpcapi.tokens"),
		},
	}
}

func main() {

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
	})

	cfg := appCfg(".")
	logrus.WithFields(logrus.Fields{
		"topic":  "config",
		"event":  "red config",
		"config": cfg,
	}).Info("running with config")

	controller := controller.Start()

	err := grpcapi.Start(&cfg.grpcapi, controller)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpcapi",
			"event": "start error",
		}).Fatal(err)
	}

	err = pcep.ListenForNewSession(controller, &cfg.pcep)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "pcep",
			"event": "ListenForNewSession error",
		}).Fatal(err)
	}
}
