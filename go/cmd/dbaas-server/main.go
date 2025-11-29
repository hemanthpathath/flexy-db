package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	"github.com/hemanthpathath/flex-db/go/internal/db"
	grpchandlers "github.com/hemanthpathath/flex-db/go/internal/grpc"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment variables
	cfg := loadConfig()

	// Connect to database
	log.Println("Connecting to database...")
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to database successfully")

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Initialize repositories
	tenantRepo := repository.NewPostgresTenantRepository(pool)
	userRepo := repository.NewPostgresUserRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	relationshipRepo := repository.NewPostgresRelationshipRepository(pool)

	// Initialize services
	tenantSvc := service.NewTenantService(tenantRepo)
	userSvc := service.NewUserService(userRepo)
	nodeTypeSvc := service.NewNodeTypeService(nodeTypeRepo)
	nodeSvc := service.NewNodeService(nodeRepo, nodeTypeRepo)
	relationshipSvc := service.NewRelationshipService(relationshipRepo, nodeRepo)

	// Initialize gRPC handlers
	tenantHandler := grpchandlers.NewTenantHandler(tenantSvc)
	userHandler := grpchandlers.NewUserHandler(userSvc)
	nodeTypeHandler := grpchandlers.NewNodeTypeHandler(nodeTypeSvc)
	nodeHandler := grpchandlers.NewNodeHandler(nodeSvc)
	relationshipHandler := grpchandlers.NewRelationshipHandler(relationshipSvc)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterTenantServiceServer(grpcServer, tenantHandler)
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterNodeTypeServiceServer(grpcServer, nodeTypeHandler)
	pb.RegisterNodeServiceServer(grpcServer, nodeHandler)
	pb.RegisterRelationshipServiceServer(grpcServer, relationshipHandler)

	// Enable reflection for grpcurl/evans
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal, stopping server...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("Starting gRPC server on port %s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// loadConfig loads database configuration from environment variables
func loadConfig() db.Config {
	cfg := db.DefaultConfig()

	if host := getEnv("DB_HOST", ""); host != "" {
		cfg.Host = host
	}
	if port := getEnv("DB_PORT", ""); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}
	if user := getEnv("DB_USER", ""); user != "" {
		cfg.User = user
	}
	if password := getEnv("DB_PASSWORD", ""); password != "" {
		cfg.Password = password
	}
	if dbName := getEnv("DB_NAME", ""); dbName != "" {
		cfg.DBName = dbName
	}
	if sslMode := getEnv("DB_SSL_MODE", ""); sslMode != "" {
		cfg.SSLMode = sslMode
	}

	return cfg
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
