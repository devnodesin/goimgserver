package server

import (
	"context"
	"fmt"
	"goimgserver/server/health"
	"goimgserver/server/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Config holds server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	EnableCORS      bool
	EnableRateLimit bool
	RateLimit       int
	RatePer         time.Duration
	Production      bool
}

// Server represents the HTTP server
type Server struct {
	Router       *gin.Engine
	httpServer   *http.Server
	config       *Config
	healthChecker *health.Checker
}

// New creates a new server with the given configuration
func New(config *Config) *Server {
	// Set Gin mode
	if config.Production {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Create router
	router := gin.New()
	
	// Set trusted proxies
	if err := router.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	}
	
	// Create server
	srv := &Server{
		Router:        router,
		config:        config,
		healthChecker: health.NewChecker(),
	}
	
	// Setup middleware
	srv.setupMiddleware()
	
	// Setup health endpoints
	srv.setupHealthEndpoints()
	
	return srv
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Request ID must be first to ensure all logs have request IDs
	s.Router.Use(middleware.RequestID())
	
	// Security headers
	s.Router.Use(middleware.SecurityHeaders())
	
	// CORS
	if s.config.EnableCORS {
		s.Router.Use(middleware.CORS())
	}
	
	// Error handling and recovery
	s.Router.Use(middleware.ErrorHandler())
	
	// Logging (after error handler to log errors too)
	s.Router.Use(middleware.Logging())
	
	// Rate limiting (if enabled)
	if s.config.EnableRateLimit {
		s.Router.Use(middleware.RateLimit(s.config.RateLimit, s.config.RatePer))
	}
}

// setupHealthEndpoints registers health check endpoints
func (s *Server) setupHealthEndpoints() {
	s.Router.GET("/health", s.healthChecker.DetailedHealthHandler)
	s.Router.GET("/live", s.healthChecker.LivenessHandler)
	s.Router.GET("/ready", s.healthChecker.ReadinessHandler)
}

// AddHealthCheck registers a health check function
func (s *Server) AddHealthCheck(name string, check health.HealthCheck) {
	s.healthChecker.AddCheck(name, check)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.Router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}
	
	log.Printf("Starting server on %s", addr)
	
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	
	log.Println("Shutting down server...")
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	
	log.Println("Server stopped gracefully")
	return nil
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run() error {
	// Start server in background
	go func() {
		if err := s.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	
	return s.Shutdown(ctx)
}
