package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/redhat-data-and-ai/usernaut/internal/httpapi/middleware"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

type APIServer struct {
	config *config.AppConfig
	router *gin.Engine
	server *http.Server
}

func NewAPIServer(cfg *config.AppConfig) *APIServer {
	if cfg.App.Environment == "local" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(&s.config.APIServer))

	s := &APIServer{
		config: cfg,
		router: router,
	}

	router.Use(middleware.CORS(&s.config.APIServer))

	s.setupRoutes()
	return s
}

func (s *APIServer) setupRoutes() {
	v1 := s.router.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth(s.config))

	v1.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "usernaut-api",
			"status":  "running",
		})
	})

	// add endpoints accordingly
}

func (s *APIServer) Start() error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.APIServer.Host, s.config.APIServer.Port),
		Handler: s.router,
	}

	go s.StopServer()
	logrus.WithField("address", s.server.Addr).Info("starting http API server")
	if err := s.server.ListenAndServe(); err != nil {
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

	if err := s.server.Shutdown(context.Background()); err != nil {
		logrus.WithError(err).Error("Error during HTTP API server shutdown")
	}

}
