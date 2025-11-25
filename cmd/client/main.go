package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ratelimitv1 "github.com/burakmert236/rate-limiter-service/api/proto"
)

func main() {
	// Parse command line flags
	key := flag.String("key", "user:123", "Rate limit key")
	limit := flag.Int("limit", 10, "Request limit")
	window := flag.Int("window", 60, "Window in seconds")
	serverAddr := flag.String("server", "localhost:50051", "Server address")
	flag.Parse()

	// Connect to gRPC server
	// Note: insecure.NewCredentials() means no TLS (OK for local development)
	conn, err := grpc.Dial(
		*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := ratelimitv1.NewRateLimiterServiceClient(conn)

	// Create request
	req := &ratelimitv1.CheckRateLimitRequest{
		Key:           *key,
		Limit:         int32(*limit),
		WindowSeconds: int32(*window),
		Namespace:     "default",
	}

	// Call the RPC method
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CheckRateLimit(ctx, req)
	if err != nil {
		log.Fatalf("CheckRateLimit failed: %v", err)
	}

	// Print response
	fmt.Println("Rate Limit Check Result:")
	fmt.Printf("  Allowed: %v\n", resp.Allowed)
	fmt.Printf("  Remaining: %d\n", resp.Remaining)
	fmt.Printf("  Limit: %d\n", resp.Limit)
	fmt.Printf("  Reset At: %v\n", resp.ResetAt.AsTime())
	if !resp.Allowed {
		fmt.Printf("  Retry After: %d seconds\n", resp.RetryAfterSeconds)
	}
}
