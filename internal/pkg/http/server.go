package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/tavsec/gin-healthcheck/config"
	timeout "github.com/vearne/gin-timeout"
	"go.uber.org/zap"
)

type Configuration struct {
	Address string `json:"address"`
}

type Server struct {
	config Configuration
	router *gin.Engine
	logger *zap.Logger
}

// NewServer creates a new HTTP server with logging and timeout middleware as well as a healthcheck endpoint.
func NewServer(logger *zap.Logger, httpCfg Configuration, checks ...checks.Check) *Server {
	router := gin.New()
	router.Use(
		ginzap.RecoveryWithZap(logger, true),
		ginzap.Ginzap(logger, time.RFC3339, true),
		timeout.Timeout(
			timeout.WithTimeout(time.Second*10),
			timeout.WithErrorHttpCode(http.StatusServiceUnavailable),
		),
	)

	// Add a healthcheck endpoint
	err := healthcheck.New(router, config.DefaultConfig(), checks)
	if err != nil {
		logger.Panic("Cannot add healthcheck endpoint", zap.Error(err))
	}

	return &Server{
		config: httpCfg,
		router: router,
	}
}

// Router exposes the router for additional endpoint configuration.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the HTTP server with the given health checks.
func (s *Server) Run(ctx context.Context) {
	server := http.Server{Addr: s.config.Address, Handler: s.router}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}
