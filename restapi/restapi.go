package restapi

import (
	"crypto/tls"
	"embed"
	"gopcep/certs"
	"gopcep/controller"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed static/*
var f embed.FS

type handler struct {
	ctr *controller.Controller
}

func (h *handler) getSessions(c *gin.Context) {
	defer h.ctr.RUnlock()

	h.ctr.RLock()
	c.JSON(200, h.ctr.PCEPSessions)
}

func (h *handler) getLSPs(c *gin.Context) {
	defer h.ctr.RUnlock()

	h.ctr.RLock()
	c.JSON(200, h.ctr.LSPs)
}

func (h *handler) getBGPNeighbors(c *gin.Context) {
	list, err := h.ctr.GetBGPNeighbor()
	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(200, list)
}

type Config struct {
	Address  string
	Port     string
	CertFile string
	KeyFile  string
}

func StartREST(cfg *Config, controller *controller.Controller) error {
	h := handler{
		ctr: controller,
	}

	router := gin.Default()

	router.Use(gin.BasicAuth(gin.Accounts{
		"someuser": "somepasss",
	}))

	// Serving static
	router.StaticFS("/ui/", http.FS(f))
	router.GET("/", func(c *gin.Context) {
		file, _ := f.ReadFile("static/index.html")
		c.Data(
			http.StatusOK,
			"text/html",
			file,
		)
	})
	router.GET("favicon.ico", func(c *gin.Context) {
		file, _ := f.ReadFile("static/favicon.ico")
		c.Data(
			http.StatusOK,
			"image/x-icon",
			file,
		)
	})

	// API methods
	apiV1 := router.Group("/v1")

	apiV1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	apiV1.GET("/pcepsessions", h.getSessions)
	apiV1.GET("/pceplsps", h.getLSPs)
	apiV1.GET("/bgpneighbors", h.getBGPNeighbors)

	// Using self signed self generated certs
	// New certs are generated during startup
	cert, pool, err := certs.GenCerts()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpc_api",
			"event": "gent certs error",
		}).Error(err)
		return err
	}

	server := http.Server{
		Addr:    cfg.Address + ":" + cfg.Port,
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
			ClientCAs:    pool,
		},
	}
	go func() {
		err = server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic": "rest_api",
				"event": "serve error",
			}).Fatal(err)
		}
	}()
	return nil
}
