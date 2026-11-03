package database

import (
	"crypto/tls"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ParseRedisConfig parses Redis address and password, supporting both URI schemes
// (redis://, rediss://) and plain host:port formats (auto-enabling TLS for Upstash).
func ParseRedisConfig(address, password string) (*redis.Options, asynq.RedisConnOpt, error) {
	trimmedAddr := strings.TrimSpace(address)
	trimmedPass := strings.TrimSpace(password)

	if strings.HasPrefix(trimmedAddr, "redis://") || strings.HasPrefix(trimmedAddr, "rediss://") {
		redisOpt, err := redis.ParseURL(trimmedAddr)
		if err != nil {
			return nil, nil, err
		}
		if trimmedPass != "" && redisOpt.Password == "" {
			redisOpt.Password = trimmedPass
		}

		asynqOpt, err := asynq.ParseRedisURI(trimmedAddr)
		if err != nil {
			return nil, nil, err
		}
		return redisOpt, asynqOpt, nil
	}

	// Plain host:port format
	addr := strings.TrimPrefix(trimmedAddr, "tcp://")
	var tlsConfig *tls.Config
	if strings.Contains(addr, "upstash.io") {
		host := strings.Split(addr, ":")[0]
		tlsConfig = &tls.Config{
			ServerName: host,
		}
	}

	redisOpt := &redis.Options{
		Addr:      addr,
		Password:  trimmedPass,
		DB:        0,
		TLSConfig: tlsConfig,
	}

	asynqOpt := asynq.RedisClientOpt{
		Addr:      addr,
		Password:  trimmedPass,
		DB:        0,
		TLSConfig: tlsConfig,
	}

	return redisOpt, asynqOpt, nil
}
