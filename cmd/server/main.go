package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ratelimitv1 "github.com/burakmert236/rate-limiter-service/api/proto"
	"github.com/burakmert236/rate-limiter-service/internal/config"
	"github.com/burakmert236/rate-limiter-service/internal/server"
	"github.com/burakmert236/rate-limiter-service/internal/storage"
)

func main() {
	cfg := config.Load()

	log.Printf("Connecting to Redis at %s...", cfg.Redis.Address)
	redisClient, err := storage.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Connected to Redis successfully")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", cfg.Server.Port, err)
	}

	grpcServer := grpc.NewServer(grpc.ConnectionTimeout(10 * time.Second))

	rateLimiterServer := server.NewServer(redisClient)
	ratelimitv1.RegisterRateLimiterServiceServer(grpcServer, rateLimiterServer)

	reflection.Register(grpcServer)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, gracefully stopping...")
		grpcServer.GracefulStop()
	}()

	log.Printf("🚀 gRPC server listening on port %d", cfg.Server.Port)
	log.Println("Press Ctrl+C to stop")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
