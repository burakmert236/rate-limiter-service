package storage

const TokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

-- Get current state from Redis
-- We store: tokens (current count) and last_refill (last update time)
local tokens = tonumber(redis.call('HGET', key, 'tokens'))
local last_refill = tonumber(redis.call('HGET', key, 'last_refill'))

-- Initialize if this is the first request for this key
if tokens == nil then
  tokens = capacity
  last_refill = now
end

-- Calculate tokens to add based on time elapsed
local time_elapsed = math.max(0, now - last_refill)
local tokens_to_add = time_elapsed * refill_rate

-- Add tokens but don't exceed capacity
tokens = math.min(capacity, tokens + tokens_to_add)

-- Update last refill time
last_refill = now

-- Check if we have enough tokens
local allowed = 0
local remaining = tokens

if tokens >= cost then
  -- We have enough tokens, consume them
  tokens = tokens - cost
  remaining = tokens
  allowed = 1
  
  -- Update Redis with new state
  redis.call('HSET', key, 'tokens', tokens, 'last_refill', last_refill)
  redis.call('EXPIRE', key, ttl)
else
  -- Not enough tokens, don't update state
  allowed = 0
end

-- Calculate when the bucket will have enough tokens
local retry_after = 0
if allowed == 0 then
  local tokens_needed = cost - tokens
  retry_after = math.ceil(tokens_needed / refill_rate)
end

-- Calculate when the bucket will be full (for reset_at)
local tokens_until_full = capacity - tokens
local seconds_until_full = math.ceil(tokens_until_full / refill_rate)

-- Return results as an array
-- [1] = allowed (1 or 0)
-- [2] = remaining tokens
-- [3] = retry_after seconds
-- [4] = seconds_until_full (for reset time)
return {allowed, remaining, retry_after, seconds_until_full}
`
