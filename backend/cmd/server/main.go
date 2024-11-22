package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/HardMax71/syncwrite/backend/pkg/auth"
	"github.com/HardMax71/syncwrite/backend/pkg/collaboration"
	"github.com/HardMax71/syncwrite/backend/pkg/config"
	"github.com/HardMax71/syncwrite/backend/pkg/database"
	"github.com/HardMax71/syncwrite/backend/pkg/document"
	"github.com/HardMax71/syncwrite/backend/pkg/health"
	authv1 "github.com/HardMax71/syncwrite/backend/pkg/proto/auth/v1"
	collaborationv1 "github.com/HardMax71/syncwrite/backend/pkg/proto/collaboration/v1"
	documentv1 "github.com/HardMax71/syncwrite/backend/pkg/proto/document/v1"
	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := utils.Logger()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize MQTT client
	mqttClient, err := collaboration.NewMQTTClient(cfg.MQTT.BrokerURL)
	if err != nil {
		logger.Fatal("Failed to connect to MQTT broker", zap.Error(err))
	}
	defer mqttClient.Close()

	// Initialize services
	authService := auth.NewService(db.Pool(), cfg)
	documentService := document.NewService(db.Pool())
	collaborationService := collaboration.NewService(db.Pool(), mqttClient)

	// Create gRPC server
	authMiddleware := auth.NewAuthMiddleware(authService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(authMiddleware.UnaryInterceptor()),
		grpc.StreamInterceptor(authMiddleware.StreamInterceptor()),
	)

	// Initialize and register health checker
	healthChecker := health.NewHealthChecker(db.Pool())
	grpc_health_v1.RegisterHealthServer(server, healthChecker)

	// Register services
	authHandler := auth.NewHandler(authService)
	documentHandler := document.NewHandler(documentService)
	collaborationHandler := collaboration.NewHandler(collaborationService, documentService)

	authv1.RegisterAuthServiceServer(server, authHandler)
	documentv1.RegisterDocumentServiceServer(server, documentHandler)
	collaborationv1.RegisterCollaborationServiceServer(server, collaborationHandler)

	// Enable reflection for development tools
	reflection.Register(server)

	// Start server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Fatal("Failed to create listener", zap.Error(err))
	}

	// Handle shutdown gracefully
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Starting gRPC server", zap.Int("port", cfg.Server.Port))
		if err := server.Serve(listener); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Clean shutdown on interrupt
	<-shutdown
	logger.Info("Shutting down server...")

	// Graceful shutdown
	server.GracefulStop()
	logger.Info("Server stopped")
}
