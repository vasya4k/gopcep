package restapi

import (
	"crypto/tls"
	"embed"
	"fmt"
	"gopcep/certs"
	"gopcep/controller"
	"net/http"

	"github.com/gin-contrib/cors"
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
	c.JSON(200, h.ctr.GetLSPs())
}

func (h *handler) getBGPNeighbors(c *gin.Context) {
	list, err := h.ctr.GetBGPNeighbor()
	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(200, list)
}

func (h *handler) createUpdRouter(c *gin.Context) {
	var r controller.Router
	err := c.BindJSON(&r)
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}
	err = h.ctr.CreateUpdRouter(&r)
	if err != nil {

		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, r)
}

func (h *handler) deleteRouter(c *gin.Context) {
	err := h.ctr.DeleteRouter(c.Param("id"))
	if err != nil {
		fmt.Println("VVVV", err)
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, c.Param("id"))
}

func (h *handler) listRouters(c *gin.Context) {
	routers, err := h.ctr.GetRouters()
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, routers)
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

	config := cors.DefaultConfig()
	config.AllowHeaders = []string{"*"}
	// Access to XMLHttpRequest at 'https://127.0.0.1:1443/v1/bgpneighbors'
	// from origin 'http://localhost:8080' has been blocked by CORS policy:
	// Response to preflight request doesn't pass access control check:
	// The value of the 'Access-Control-Allow-Origin' header in the
	// response must not be the wildcard '*' when the request's credentials mode is 'include'.
	// The credentials mode of requests initiated by the XMLHttpRequest is controlled by the withCredentials attribute.
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"*"}
	config.ExposeHeaders = []string{"*"}
	//Need cors if UI is served from a different server
	router.Use(cors.New(config))
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
	// router methods
	apiV1.POST("/router", h.createUpdRouter)
	apiV1.DELETE("/router/:id", h.deleteRouter)
	apiV1.GET("/routers", h.listRouters)

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
