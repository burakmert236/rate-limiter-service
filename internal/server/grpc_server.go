package server

import (
	"context"
	"fmt"
	"log"

	ratelimitv1 "github.com/burakmert236/rate-limiter-service/api/proto"
	"github.com/burakmert236/rate-limiter-service/internal/ratelimiter"
	"github.com/burakmert236/rate-limiter-service/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	ratelimitv1.UnimplementedRateLimiterServiceServer

	tokenBucket *ratelimiter.TokenBucketLimiter

	redisClient *storage.RedisClient
}

func NewServer(redisClient *storage.RedisClient) *Server {
	return &Server{
		tokenBucket: ratelimiter.NewTokenBucketLimiter(redisClient),
		redisClient: redisClient,
	}
}

func (s *Server) CheckRateLimit(
	ctx context.Context,
	request *ratelimitv1.CheckRateLimitRequest,
) (*ratelimitv1.CheckRateLimitResponse, error) {

	if request.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	if request.Limit <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be positive")
	}
	if request.WindowSeconds <= 0 {
		return nil, status.Error(codes.InvalidArgument, "window_seconds must be positive")
	}

	fullKey := request.Key
	if request.Namespace != "" {
		fullKey = fmt.Sprintf("%s:%s", request.Namespace, request.Key)
	}

	allowed, remaining, resetAt, retryAfter, err := s.tokenBucket.AllowRequest(
		ctx,
		fullKey,
		request.Limit,
		request.WindowSeconds,
	)

	if err != nil {
		log.Printf("Rate limit check error: %v", err)
		return nil, status.Errorf(codes.Internal, "rate limit check failed: %v", err)
	}

	log.Printf("Rate limit check: key=%s, allowed=%v, remaining=%d, limit=%d",
		fullKey, allowed, remaining, request.Limit)

	return &ratelimitv1.CheckRateLimitResponse{
		Allowed:           allowed,
		Remaining:         remaining,
		ResetAt:           timestamppb.New(resetAt),
		RetryAfterSeconds: retryAfter,
		Limit:             request.Limit,
	}, nil
}

func (s *Server) HealthCheck(ctx context.Context) error {
	return s.redisClient.Ping(ctx)
}
