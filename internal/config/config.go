package config

import (
	"time"
)

type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

type ServerConfig struct {
	Port int
}

type Config struct {
	Redis  RedisConfig
	Server ServerConfig
}

func Load() *Config {
	return &Config{
		Redis: RedisConfig{
			Address:      "localhost:6379",
			Password:     "",
			DB:           0,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		},
		Server: ServerConfig{
			Port: 50051,
		},
	}
}
