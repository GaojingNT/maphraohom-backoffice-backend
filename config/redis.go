package config

import (
	"os"
	"strconv"
)

type redisConfig struct {
	RedisHost          string
	RedisPort          string
	RedisUsername      string
	RedisPassword      string
	RedisCachePrefix   string
	RedisCacheDuration int
}

func NewRedisConfig() *redisConfig {
	redisCacheDuration := func() int {
		// Default max cache duration is 5
		redisCacheDuration := 5
		envRedisCacheDuration, err := strconv.Atoi(os.Getenv("REDIS_CACHE_DURATION"))
		if err == nil {
			redisCacheDuration = envRedisCacheDuration
		}
		return redisCacheDuration
	}()

	return &redisConfig{
		RedisHost:          os.Getenv("REDIS_HOST"),
		RedisPort:          os.Getenv("REDIS_PORT"),
		RedisUsername:      os.Getenv("REDIS_USERNAME"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisCachePrefix:   os.Getenv("REDIS_CACHE_PREFIX"),
		RedisCacheDuration: redisCacheDuration,
	}
}
