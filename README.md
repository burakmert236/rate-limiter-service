# Rate Limiter Service

A high-performance, distributed rate limiting service built with Go, gRPC, and Redis. Implements the Token Bucket algorithm to provide accurate and efficient rate limiting for microservices architectures.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![gRPC](https://img.shields.io/badge/gRPC-latest-244c5a?style=flat&logo=google)](https://grpc.io/)
[![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)

## ✨ Features

- **High Performance**: Built with Go for excellent concurrency and low latency
- **Token Bucket Algorithm**: Accurate rate limiting with burst traffic support
- **Distributed**: Redis-backed for horizontal scalability
- **gRPC API**: Modern, efficient communication protocol
- **Atomic Operations**: Lua scripts ensure consistency in concurrent environments
- **Multi-tenancy**: Namespace support for organizing rate limits
- **Flexible Configuration**: Environment variable-based configuration
- **Production-Ready**: Docker support, health checks, and graceful shutdown
- **Type-Safe**: Protocol Buffers for strongly-typed API contracts

## 🏗 Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Client    │  gRPC   │  Rate Limiter    │  Lua    │    Redis    │
│ Application │────────▶│     Service      │────────▶│   Cluster   │
│             │         │  (Go + gRPC)     │  Script │             │
└─────────────┘         └──────────────────┘         └─────────────┘
                               │
                               │ Token Bucket Algorithm
                               │
                        ┌──────▼─────────┐
                        │ Capacity: 10   │
                        │ Rate: 1/sec    │
                        │ Tokens: 7      │
                        └────────────────┘
```

**Key Components:**

- **gRPC Server**: Handles incoming rate limit check requests
- **Token Bucket Limiter**: Implements rate limiting logic
- **Redis Client**: Manages distributed state with atomic operations
- **Lua Scripts**: Ensures atomic token bucket operations in Redis

**How it Works:**

1. Client sends rate limit check request via gRPC
2. Server calculates token refill based on elapsed time
3. Lua script atomically checks/updates tokens in Redis
4. Response indicates if request is allowed + metadata
5. Tokens refill continuously at configured rate

## 🔧 Prerequisites

- **Go** 1.23 or higher
- **Redis** 7.0 or higher
- **Docker** (optional, for containerized deployment)
- **protoc** (Protocol Buffer compiler)
- **grpcurl** (optional, for testing)

### Installing Prerequisites on Windows

```powershell
# Install Go
# Download from: https://go.dev/dl/

# Install Chocolatey (if not already installed)
Set-ExecutionPolicy Bypass -Scope Process -Force
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install protoc and grpcurl
choco install protoc grpcurl

# Install Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/rate-limiter-service.git
cd rate-limiter-service
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Generate gRPC Code from Proto Files

```bash
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File scripts/generate-proto.ps1

# Or using Make (if installed)
make proto
```

### 4. Start Redis

```bash
# Using Docker Compose
cd deployments
docker-compose up -d redis

# Or using Docker directly
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

### 5. Run the Service

```bash
# From project root
go run cmd/server/main.go
```

You should see:
```
Connecting to Redis at localhost:6379...
✓ Connected to Redis successfully
🚀 gRPC server listening on port 50051
Press Ctrl+C to stop
```

### 6. Test with CLI Client

Open a new terminal:

```bash
# Allow a request
go run cmd/client/main.go --key "user:123" --limit 10 --window 60

# Output:
# Rate Limit Check Result:
#   Allowed: true
#   Remaining: 9
#   Limit: 10
#   Reset At: 2024-01-15 10:30:45 +0000 UTC
```

## 📚 Usage

### Using the Go Client

```bash
go run cmd/client/main.go \
  --key "user:123" \
  --limit 10 \
  --window 60 \
  --server "localhost:50051"
```

**Parameters:**
- `--key`: Unique identifier for rate limiting (e.g., user ID, API key, IP address)
- `--limit`: Maximum number of requests allowed
- `--window`: Time window in seconds
- `--server`: Server address (default: localhost:50051)

### Using grpcurl

```bash
grpcurl -plaintext -d '{
  "key": "user:456",
  "limit": 100,
  "window_seconds": 3600,
  "namespace": "api"
}' localhost:50051 ratelimit.v1.RateLimitService/CheckRateLimit
```

### Example Response

```json
{
  "allowed": true,
  "remaining": 99,
  "resetAt": "2024-01-15T11:30:45Z",
  "retryAfterSeconds": 0,
  "limit": 100
}
```

### Common Use Cases

**Per-User Rate Limiting:**
```bash
# Allow 100 requests per hour per user
--key "user:${USER_ID}" --limit 100 --window 3600
```

**Per-IP Rate Limiting:**
```bash
# Allow 1000 requests per day per IP
--key "ip:${IP_ADDRESS}" --limit 1000 --window 86400
```

**Per-API-Key Rate Limiting:**
```bash
# Allow 10,000 requests per hour per API key
--key "apikey:${API_KEY}" --limit 10000 --window 3600
```

**Per-Endpoint Rate Limiting:**
```bash
# Allow 50 requests per minute per endpoint
--key "endpoint:/api/users" --limit 50 --window 60
```

**Composite Keys:**
```bash
# Allow 10 requests per minute per user per endpoint
--key "user:${USER_ID}:endpoint:/api/upload" --limit 10 --window 60
```

## ⚙️ Configuration

The service is configured via environment variables:

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `50051` | gRPC server port |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |

### Redis Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | `` | Redis password (if required) |
| `REDIS_DB` | `0` | Redis database number (0-15) |
| `REDIS_POOL_SIZE` | `10` | Maximum connection pool size |
| `REDIS_MIN_IDLE_CONNS` | `2` | Minimum idle connections |
| `REDIS_MAX_RETRIES` | `3` | Maximum retry attempts |
| `REDIS_DIAL_TIMEOUT` | `5s` | Connection timeout |
| `REDIS_READ_TIMEOUT` | `3s` | Read operation timeout |
| `REDIS_WRITE_TIMEOUT` | `3s` | Write operation timeout |

### Example Configuration

```bash
# .env file
SERVER_PORT=50051
LOG_LEVEL=debug

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=20
```

## 📖 API Reference

### RateLimitService

#### CheckRateLimit

Checks if a request should be allowed under the rate limit.

**Request:**
```protobuf
message CheckRateLimitRequest {
  string key = 1;              // Unique identifier
  int32 limit = 2;             // Maximum requests allowed
  int32 window_seconds = 3;    // Time window in seconds
  string namespace = 4;        // Optional namespace
}
```

**Response:**
```protobuf
message CheckRateLimitResponse {
  bool allowed = 1;                          // Whether request is allowed
  int32 remaining = 2;                       // Tokens remaining
  google.protobuf.Timestamp reset_at = 3;    // When bucket will be full
  int32 retry_after_seconds = 4;             // Wait time if denied
  int32 limit = 5;                           // Current limit
}
```

## 🔨 Development

### Project Setup

```bash
# Clone repository
git clone https://github.com/burakmert236/rate-limiter-service.git
cd rate-limiter-service

# Install dependencies
go mod download

# Generate proto files
make proto  # or use scripts/generate-proto.ps1

# Run with hot reload (using air)
go install github.com/cosmtrek/air@latest
air
```

### Regenerating Proto Files

After modifying `.proto` files:

```bash
# Windows
powershell -ExecutionPolicy Bypass -File scripts/generate-proto.ps1

# Linux/Mac
make proto
```

## 🧪 Testing

### Manual Testing

```bash
# Test with different scenarios
go run cmd/client/main.go --key "test:1" --limit 5 --window 10

# Rapid fire test (exhaust limit)
for i in {1..10}; do
  go run cmd/client/main.go --key "test:2" --limit 5 --window 60
done

# Test with grpcurl
grpcurl -plaintext -d '{"key":"test:3","limit":10,"window_seconds":60}' \
  localhost:50051 ratelimit.v1.RateLimitService/CheckRateLimit
```

## 🐳 Docker

### Running with Docker Compose

```bash
cd deployments

# Start all services (Redis + Rate Limiter)
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```


### Running Standalone Containers

```bash
# Run Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Run Rate Limiter
docker run -d --name rate-limiter \
  -p 50051:50051 \
  -e REDIS_ADDR=host.docker.internal:6379 \
  rate-limiter-service:latest

# View logs
docker logs -f rate-limiter

# Stop containers
docker stop rate-limiter redis
docker rm rate-limiter redis
```
