package restapi

import (
	"crypto/tls"
	"embed"
	"gopcep/certs"
	"gopcep/controller"
	"net/http"
	"strings"
	"time"

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
	Debug    bool
}

type handler struct {
	ctr *controller.Controller
}

// GetClientIP gets the correct IP for the end client instead of the proxy
func GetClientIP(c *gin.Context) string {
	// first check the X-Forwarded-For header
	requester := c.Request.Header.Get("X-Forwarded-For")
	// if empty, check the Real-IP header
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	// if the requester is still empty, use the hard-coded address from the socket
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	// if requester is a comma delimited list, take the first one
	// (this happens when proxied via elastic load balancer then again through nginx)
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}

	return requester
}

// jsonLogMiddleware logs a gin HTTP request in JSON format, with some additional custom key/values
func jsonLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// process request
		c.Next()

		entry := logrus.WithFields(logrus.Fields{
			"client_ip": GetClientIP(c),
			"duration": func() float64 {
				milliseconds := float64(time.Since(start)) / float64(time.Millisecond)
				return float64(milliseconds*100+.5) / 100
			}(),
			"method":     c.Request.Method,
			"path":       c.Request.RequestURI,
			"status":     c.Writer.Status(),
			"referrer":   c.Request.Referer(),
			"request_id": c.Writer.Header().Get("Request-Id"),
		})

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("")
		}
	}
}

func (h *handler) getBGPNeighbors(c *gin.Context) {
	list, err := h.ctr.GetBGPNeighbor()
	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(200, list)
}

func newCORSMidleware(cfg *Config) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowHeaders = []string{"*"}
	config.AllowCredentials = true
	// Access to XMLHttpRequest at 'https://127.0.0.1:1443/v1/bgpneighbors'
	// from origin 'http://localhost:8080' has been blocked by CORS policy:
	// Response to preflight request doesn't pass access control check:
	// The value of the 'Access-Control-Allow-Origin' header in the
	// response must not be the wildcard '*' when the request's credentials mode is 'include'.
	// The credentials mode of requests initiated by the XMLHttpRequest is controlled by the withCredentials attribute.
	config.AllowOrigins = []string{"http://localhost:8080", "http://localhost:8082", "https://localhost:" + cfg.Port}
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

func Start(cfg *Config, controller *controller.Controller) error {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	//Need cors if UI is served from a different server
	router.Use(newCORSMidleware(cfg))
	// Add basic auth
	router.Use(gin.BasicAuth(gin.Accounts{
		cfg.User: cfg.Pass,
	}))

	router.Use(jsonLogMiddleware())

	// /ui/static/
	// Serving static from embedded file system
	// router.StaticFS("static/", http.FS(f))
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
