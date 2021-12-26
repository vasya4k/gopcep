package restapi

import (
	"crypto/tls"
	"embed"
	"gopcep/certs"
	"gopcep/controller"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed static/*
var f embed.FS

type Config struct {
	Address  string
	Port     string
	CertFile string
	KeyFile  string
	User     string
	Pass     string
}

type handler struct {
	ctr *controller.Controller
}

func (h *handler) getBGPNeighbors(c *gin.Context) {
	list, err := h.ctr.GetBGPNeighbor()
	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(200, list)
}

func newCORSMidleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowHeaders = []string{"*"}
	// Access to XMLHttpRequest at 'https://127.0.0.1:1443/v1/bgpneighbors'
	// from origin 'http://localhost:8080' has been blocked by CORS policy:
	// Response to preflight request doesn't pass access control check:
	// The value of the 'Access-Control-Allow-Origin' header in the
	// response must not be the wildcard '*' when the request's credentials mode is 'include'.
	// The credentials mode of requests initiated by the XMLHttpRequest is controlled by the withCredentials attribute.
	config.AllowOrigins = []string{"http://localhost:8080", "http://localhost:8082"}
	config.AllowMethods = []string{"*"}
	config.ExposeHeaders = []string{"*"}
	return cors.New(config)
}

func newServer(router *gin.Engine, cfg *Config) (*http.Server, error) {
	// Using self signed self generated certs
	// New certs are generated during startup
	cert, pool, err := certs.GenCerts()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "grpc_api",
			"event": "gent certs error",
		}).Error(err)
		return nil, err
	}

	server := http.Server{
		Addr:    cfg.Address + ":" + cfg.Port,
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
			ClientCAs:    pool,
		},
	}
	return &server, nil
}

func addAPIMethods(router *gin.Engine, controller *controller.Controller) {
	h := handler{
		ctr: controller,
	}
	// API methods
	apiV1 := router.Group("/v1")

	apiV1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// PCEP
	apiV1.GET("/pcepsessions", h.getSessions)
	// BGP
	apiV1.GET("/bgpneighbors", h.getBGPNeighbors)
	// Router methods
	apiV1.POST("/router", h.createUpdRouter)
	apiV1.DELETE("/router/:id", h.deleteRouter)
	apiV1.GET("/routers", h.listRouters)
	// LSP methods
	apiV1.POST("/lsp", h.createUpdLSP)
	apiV1.DELETE("/lsp/:name", h.delLSP)
	apiV1.GET("/pceplsps", h.getLSPs)
	apiV1.GET("/ctrlsps", h.getNetLSPs)
}

func StartREST(cfg *Config, controller *controller.Controller) error {
	router := gin.Default()
	//Need cors if UI is served from a different server
	router.Use(newCORSMidleware())
	// Add basic auth
	router.Use(gin.BasicAuth(gin.Accounts{
		cfg.User: cfg.Pass,
	}))
	// Serving static from embeded file system
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

	addAPIMethods(router, controller)

	server, err := newServer(router, cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "rest_api",
			"event": "new server error",
		}).Fatal(err)
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
