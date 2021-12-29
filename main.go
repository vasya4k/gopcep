package main

import (
	"gopcep/controller"
	"gopcep/grpcapi"
	"gopcep/pcep"
	"gopcep/restapi"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

type logCfg struct {
	FullTimestamp   bool
	DisableColors   bool
	TimestampFormat string
	TextFormat      bool
	LogLevel        uint32
}

type cfg struct {
	grpcapi grpcapi.Config
	restapi restapi.Config
	pcep    pcep.Cfg
	logCfg  logCfg
	bgpls   controller.BGPGlobalCfg
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
		},
		grpcapi: grpcapi.Config{
			ListenAddr: viper.GetString("grpcapi.listen_addr"),
			ListenPort: viper.GetString("grpcapi.listen_port"),
			Tokens:     viper.GetStringSlice("grpcapi.tokens"),
		},
		bgpls: controller.BGPGlobalCfg{
			AS:       uint32(viper.GetUint32("bgpls.as")),
			RouterId: viper.GetString("bgpls.router_id"),
		},
		restapi: restapi.Config{
			Address:  viper.GetString("restapi.listen_addr"),
			Port:     viper.GetString("restapi.listen_port"),
			CertFile: viper.GetString("restapi.cert_file"),
			KeyFile:  viper.GetString("restapi.key_file"),
			User:     viper.GetString("restapi.user"),
			Pass:     viper.GetString("restapi.pass"),
			Debug:    viper.GetBool("restapi.debug"),
		},
		logCfg: logCfg{
			LogLevel:        viper.GetUint32("log.level"),
			TimestampFormat: viper.GetString("log.time_format"),
			TextFormat:      viper.GetBool("log.text_format"),
			FullTimestamp:   viper.GetBool("log.full_timestamp"),
			DisableColors:   viper.GetBool("log.disable_colors"),
		},
	}
}

func configureLogging(cfg logCfg) {
	logrus.SetLevel(logrus.Level(cfg.LogLevel))

	if cfg.TextFormat {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:   cfg.DisableColors,
			FullTimestamp:   cfg.FullTimestamp,
			TimestampFormat: cfg.TimestampFormat,
		})
		return
	}
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: cfg.TimestampFormat,
	})
}

func startController(c *cli.Context) error {
	cfg := appCfg(c.String("config"))
	configureLogging(cfg.logCfg)

	logrus.WithFields(logrus.Fields{
		"topic":  "config",
		"event":  "red config",
		"config": cfg,
	}).Info("running with config")

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "db",
			"event": "failed to open db",
		}).Fatal(err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic": "db",
				"event": "failed to close db",
			}).Error(err)
		}
	}()

	controller := controller.Start(db, &cfg.bgpls)

	err = grpcapi.Start(&cfg.grpcapi, controller)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpcapi",
			"event": "start error",
		}).Fatal(err)
	}

	err = restapi.Start(&cfg.restapi, controller)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "restapi",
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
	return nil
}

func main() {
	app := &cli.App{
		Name:    "GoPCEP",
		Usage:   "Segment Routing Traffic Engineering Controller written in Go",
		Version: "0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config, c",
				Value: ".",
				Usage: "config path",
			},
		},
		Action: startController,
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "run",
			"event": "failed to start",
		}).Fatal(err)
	}
}
