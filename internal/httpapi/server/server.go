package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/redhat-data-and-ai/usernaut/internal/httpapi/handlers"
	"github.com/redhat-data-and-ai/usernaut/internal/httpapi/middleware"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

type APIServer struct {
	config   *config.AppConfig
	router   *gin.Engine
	server   *http.Server
	handlers *handlers.Handlers
}

func NewAPIServer(cfg *config.AppConfig, dataStore *store.Store) *APIServer {
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"error":      param.ErrorMessage,
		}).Info("HTTP request")
		return ""
	}))
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(&cfg.APIServer))

	s := &APIServer{
		config:   cfg,
		router:   router,
		handlers: handlers.NewHandlers(cfg, dataStore),
	}

	s.setupRoutes()
	return s
}

func (s *APIServer) setupRoutes() {

	s.router.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "usernaut-api",
			"status":  "running",
		})
	})

	v1 := s.router.Group("/api/v1")
	v1.Use(middleware.BasicAuth(s.config))

	// add authenticated endpoints accordingly

	v1.GET("/backends", s.handlers.GetBackends)
	v1.GET("/user/:email/groups", s.handlers.GetUserGroups)

}

func (s *APIServer) Start() error {
	s.server = &http.Server{
		Addr:    s.config.APIServer.Address,
		Handler: s.router,
	}

	go s.StopServer()

	logrus.WithField("address", s.server.Addr).Info("starting http API server")
	if err := s.server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			logrus.Info("http API server stopped")
			return nil
		}
		return fmt.Errorf("failed to start http API server : %w", err)
	}

	return nil
}

func (s *APIServer) StopServer() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("turning down http API server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("error during HTTP API server shutdown")
	}

}
