# PowerShell script to generate Go code from proto files

# Colors for output
$ErrorColor = "Red"
$SuccessColor = "Green"
$InfoColor = "Blue"

Write-Host "Generating Go code from proto files..." -ForegroundColor $InfoColor

# Generate Go code
protoc `
  --go_out=. `
  --go_opt=paths=source_relative `
  --go-grpc_out=. `
  --go-grpc_opt=paths=source_relative `
  api/proto/ratelimit.proto