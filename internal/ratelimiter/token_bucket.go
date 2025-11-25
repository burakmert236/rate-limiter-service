package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/burakmert236/rate-limiter-service/internal/storage"
	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	redisClient *storage.RedisClient
	script      *redis.Script
}

func NewTokenBucketLimiter(redisClient *storage.RedisClient) *TokenBucketLimiter {
	script := redis.NewScript(storage.TokenBucketScript)

	return &TokenBucketLimiter{
		redisClient: redisClient,
		script:      script,
	}
}

func (tb *TokenBucketLimiter) AllowRequest(
	ctx context.Context,
	key string,
	limit int32,
	windowSeconds int32,
) (allowed bool, remaining int32, resetAt time.Time, retryAfter int32, err error) {

	if key == "" {
		return false, 0, time.Time{}, 0, fmt.Errorf("key cannot be empty")
	}
	if limit <= 0 {
		return false, 0, time.Time{}, 0, fmt.Errorf("limit must be positive")
	}
	if windowSeconds <= 0 {
		return false, 0, time.Time{}, 0, fmt.Errorf("windowSeconds must be positive")
	}

	capacity := float64(limit)
	refillRate := float64(limit) / float64(windowSeconds)
	now := time.Now()
	nowUnix := float64(now.Unix())
	cost := 1.0
	ttl := windowSeconds * 2

	redisKey := GetRedisKey(key)

	result, err := tb.script.Run(
		ctx,
		tb.redisClient.GetClient(),
		[]string{redisKey}, // KEYS
		capacity,           // ARGV[1]
		refillRate,         // ARGV[2]
		nowUnix,            // ARGV[3]
		cost,               // ARGV[4]
		ttl,                // ARGV[5]
	).Result()

	if err != nil {
		return false, 0, time.Time{}, 0, fmt.Errorf("redis script execution failed: %w", err)
	}

	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) != 4 {
		return false, 0, time.Time{}, 0, fmt.Errorf("unexpected script result format")
	}

	allowedInt, _ := resultArray[0].(int64)
	remainingFloat, _ := resultArray[1].(string)
	retryAfterInt, _ := resultArray[2].(int64)
	secondsUntilFullInt, _ := resultArray[3].(int64)

	allowed = allowedInt == 1

	var remainingTokens float64
	fmt.Sscanf(remainingFloat, "%f", &remainingTokens)
	remaining = int32(remainingTokens)

	retryAfter = int32(retryAfterInt)
	resetAt = now.Add(time.Duration(secondsUntilFullInt) * time.Second)

	return allowed, remaining, resetAt, retryAfter, nil
}

func (tb *TokenBucketLimiter) GetCurrentTokens(
	ctx context.Context,
	key string,
) (tokens float64, lastRefill time.Time, err error) {

	redisKey := GetRedisKey(key)

	// Get current state from Redis
	result, err := tb.redisClient.GetClient().HGetAll(ctx, redisKey).Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to get tokens: %w", err)
	}

	if len(result) == 0 {
		// Key doesn't exist yet
		return 0, time.Time{}, nil
	}

	// Parse stored values
	fmt.Sscanf(result["tokens"], "%f", &tokens)
	var lastRefillUnix float64
	fmt.Sscanf(result["last_refill"], "%f", &lastRefillUnix)
	lastRefill = time.Unix(int64(lastRefillUnix), 0)

	return tokens, lastRefill, nil
}

func (tb *TokenBucketLimiter) Reset(ctx context.Context, key string) error {
	redisKey := GetRedisKey(key)
	return tb.redisClient.GetClient().Del(ctx, redisKey).Err()
}

func GetRedisKey(key string) string {
	return fmt.Sprintf("ratelimit:token_bucket:%s", key)
}
