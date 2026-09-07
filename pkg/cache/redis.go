package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
)

var AppCacher *Cache

func CurrentCacher() *Cache {
	return AppCacher
}

type (
	Cache struct {
		Redis  *redis.Client
		logger *logger.Logger
		// redisCluster *redis.ClusterClient
		tags []string

		prefix  string
		expired time.Duration
	}

	// Cache service functions
	ServicePaginationFunc                                  func(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error)
	ServiceQueryByParamFunc[P interface{}, R interface{}]  func(ctx context.Context, param P) (R, error)
	ServiceQueryByParamsFunc[P interface{}, R interface{}] func(ctx context.Context, params ...P) (R, error)
	ServiceQueryFunc[R interface{}]                        func(ctx context.Context) (R, error)
)

func Initialize() *redis.Client {
	var redisClient *redis.Client

	if config.Global.Redis.RedisHost != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.Global.Redis.RedisHost, config.Global.Redis.RedisPort),
			Username: config.Global.Redis.RedisUsername,
			Password: config.Global.Redis.RedisPassword,
		})

		return redisClient
	}

	return nil
}

func NewCacher(redisClient *redis.Client, opts ...Option) *Cache {
	var (
		resultPing string
		err        error
	)

	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	// No Redis
	if redisClient == nil {
		if !fiber.IsChild() {
			log.Println("[App] Redis: client connection is", color.Format(color.RED, "off!"))
		}

		return &Cache{
			Redis:   nil,
			logger:  logger.AppLogger,
			prefix:  o.prefix,
			expired: o.expired,
		}
	}

	if !fiber.IsChild() {
		log.Println("[App] Connecting to Redis server...")
	}

	utils.Block{
		Try: func() {
			resultPing, err = redisClient.Ping(context.Background()).Result()
			if err != nil {
				utils.Throw(err)
			}

			if !fiber.IsChild() {
				if resultPing != "" {
					log.Println("[App] Redis client connected", color.Format(color.GREEN, "successfully!"))
				}
			}
		},
		Catch: func(e utils.Exception) {
			if !fiber.IsChild() {
				logger.AppLogger.Warn(e.(error).Error())
				log.Println("[App] Redis: client connection is", color.Format(color.RED, "off!"))
			}
		},
		Finally: nil,
	}.Do()

	return &Cache{
		Redis:   redisClient,
		logger:  logger.AppLogger,
		prefix:  o.prefix,
		expired: o.expired,
	}
}

type Options struct {
	prefix  string
	expired time.Duration
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.prefix = prefix
	}
}

func WithExpired(exp time.Duration) Option {
	return func(o *Options) {
		o.expired = exp
	}
}

func (c *Cache) Tag(tag ...string) *Cache {
	c.tags = tag
	return c
}

func (c *Cache) Get(ctx context.Context, key string, val interface{}) error {
	if len(c.prefix) > 0 {
		key = c.prefix + ":" + key
	}

	// Case: No Redis
	if c.Redis == nil {
		return nil
	}

	jsonStr, err := c.Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}

		// Case: Connection refused then ignore logging message
		if !strings.Contains(err.Error(), "connection refused") {
			c.logger.Error(fmt.Sprintf("c.redis.Get err: %s\n", err.Error()))
		}

		return err
	}

	err = json.Unmarshal([]byte(jsonStr), &val)
	if err != nil {
		c.logger.Error(fmt.Sprintf("c.redis.Get json.Unmarshal err: %s\n", err.Error()))
		return err
	}

	return nil
}

func (c *Cache) Set(ctx context.Context, key string, val interface{}) error {
	if len(c.prefix) > 0 {
		key = c.prefix + ":" + key
	}

	// Case: No Redis
	if c.Redis == nil {
		return nil
	}

	_, err := c.Redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		for _, v := range c.tags {
			err := p.SAdd(ctx, c.prefix+":"+v, key).Err()
			if err != nil {
				c.logger.Error(fmt.Sprintf("c.redis.Set p.SAdd err: %s\n", err.Error()))
				return err
			}
		}

		value, err := json.Marshal(val)
		if err != nil {
			c.logger.Error(fmt.Sprintf("c.redis.Set json.Unmarshal err: %s\n", err.Error()))
			return err
		}

		err = p.Set(ctx, key, string(value), c.expired).Err()
		if err != nil {
			c.logger.Error(fmt.Sprintf("c.redis.Set p.Set err: %s\n", err.Error()))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Flush(ctx context.Context) error {
	// Case: No Redis
	if c.Redis == nil {
		return nil
	}

	_, err := c.Redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		for _, v := range c.tags {
			members, err := c.Redis.SMembers(ctx, c.prefix+":"+v).Result()
			if err != nil {
				c.logger.Error(fmt.Sprintf("c.redis.Set p.Set err: %s\n", err.Error()))
				fmt.Println("c.redis.SMembers err:", err)
				return err
			}

			if len(members) > 0 {
				err = p.Del(ctx, members...).Err()
				if err != nil {
					c.logger.Error(fmt.Sprintf("c.redis.Flush p.Del err: %s\n", err.Error()))
					return err
				}
			}

			err = p.Del(ctx, c.prefix+":"+v).Err()
			if err != nil {
				c.logger.Error(fmt.Sprintf("c.redis.Flush p.Del err: %s\n", err.Error()))
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Close() {
	// Case: No Redis
	if c.Redis == nil {
		return
	}

	c.Redis.Close()
}

// Cache methods

func (c *Cache) Indexing(tags []string, keyName string, opts ...interface{}) ([]string, string) {
	cacheTags := tags
	cacheKey := keyName

	for _, opt := range opts {
		cacheKey = fmt.Sprintf(`%s_%v`, cacheKey, opt)
	}

	return cacheTags, cacheKey
}

func (c *Cache) PaginationCache(ctx context.Context, key string, tags []string, paginate *paginator.Pagination, f ServicePaginationFunc) (*paginator.Pagination, error) {
	var (
		responseData *paginator.Pagination
		err          error
	)

	if c.Redis != nil {
		// Get the cached attributes object
		err = c.Get(ctx, key, &responseData)
		if err != nil {
			// Note: If the cache is not found, it will return an error
			// Then, it will call the service function by default
			// Call service function
			responseData, err = f(ctx, paginate)
			if err != nil {
				return nil, err
			}

			// Force return the data
			return responseData, nil
		}

		if responseData == nil {
			// Call service function
			responseData, err = f(ctx, paginate)
			if err != nil {
				return nil, err
			}

			// Set cache
			err = c.Tag(tags...).Set(ctx, key, &responseData)
			if err != nil {
				return nil, err
			}
		}

		return responseData, nil
	}

	// Call service function
	responseData, err = f(ctx, paginate)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func QueryByParamCache[P, R any](c *Cache, ctx context.Context, key string, tags []string, param P, f ServiceQueryByParamFunc[P, R]) (R, error) {
	var (
		responseData R
		cacheData    map[string]interface{}
		err          error
	)

	if c.Redis != nil {
		// Get the cached attributes object
		err = c.Get(ctx, key, &cacheData)
		if err != nil {
			// Note: If the cache is not found, it will return an error
			// Then, it will call the service function by default
			// Call service function
			responseData, err = f(ctx, param)
			if err != nil {
				return responseData, err
			}

			// Force return the data
			return responseData, nil
		}

		// Check if cache data is not nil
		if len(cacheData) == 0 {
			// Call service function
			responseData, err = f(ctx, param)
			if err != nil {
				return responseData, err
			}

			// Set cache
			err = c.Tag(tags...).Set(ctx, key, &responseData)
			if err != nil {
				return responseData, err
			}

			return responseData, nil
		}

		// Convert cache data to bytes
		cacheBytes, err := json.Marshal(cacheData)
		if err != nil {
			return responseData, err
		}

		// Unmarshal cache data
		err = json.Unmarshal(cacheBytes, &responseData)
		if err != nil {
			return responseData, err
		}

		return responseData, nil
	}

	// Call service function
	responseData, err = f(ctx, param)
	if err != nil {
		return responseData, err
	}

	return responseData, nil
}

func QueryByParamsCache[P, R any](c *Cache, ctx context.Context, key string, tags []string, f ServiceQueryByParamsFunc[P, R], params ...P) (R, error) {
	var (
		responseData R
		cacheData    map[string]interface{}
		err          error
	)

	if c.Redis != nil {
		// Get the cached attributes object
		err = c.Get(ctx, key, &cacheData)
		if err != nil {
			// Note: If the cache is not found, it will return an error
			// Then, it will call the service function by default
			// Call service function
			responseData, err = f(ctx, params...)
			if err != nil {
				return responseData, err
			}

			// Force return the data
			return responseData, nil
		}

		// Check if cache data is not nil
		if len(cacheData) == 0 {
			// Call service function
			responseData, err = f(ctx, params...)
			if err != nil {
				return responseData, err
			}

			// Set cache
			err = c.Tag(tags...).Set(ctx, key, &responseData)
			if err != nil {
				return responseData, err
			}

			return responseData, nil
		}

		// Convert cache data to bytes
		cacheBytes, err := json.Marshal(cacheData)
		if err != nil {
			return responseData, err
		}

		// Unmarshal cache data
		err = json.Unmarshal(cacheBytes, &responseData)
		if err != nil {
			return responseData, err
		}

		return responseData, nil
	}

	// Call service function
	responseData, err = f(ctx, params...)
	if err != nil {
		return responseData, err
	}

	return responseData, nil
}

func QueryCache[R any](c *Cache, ctx context.Context, key string, tags []string, f ServiceQueryFunc[R]) (R, error) {
	var (
		responseData R
		cacheData    map[string]interface{}
		err          error
	)

	if c.Redis != nil {
		// Get the cached attributes object
		err = c.Get(ctx, key, &cacheData)
		if err != nil {
			// Note: If the cache is not found, it will return an error
			// Then, it will call the service function by default
			// Call service function
			responseData, err = f(ctx)
			if err != nil {
				return responseData, err
			}

			// Force return the data
			return responseData, nil
		}

		// Check if cache data is not nil
		if len(cacheData) == 0 {
			// Call service function
			responseData, err = f(ctx)
			if err != nil {
				return responseData, err
			}

			// Set cache
			err = c.Tag(tags...).Set(ctx, key, &responseData)
			if err != nil {
				return responseData, err
			}

			return responseData, nil
		}

		// Convert cache data to bytes
		cacheBytes, err := json.Marshal(cacheData)
		if err != nil {
			return responseData, err
		}

		// Unmarshal cache data
		err = json.Unmarshal(cacheBytes, &responseData)
		if err != nil {
			return responseData, err
		}

		return responseData, nil
	}

	// Call service function
	responseData, err = f(ctx)
	if err != nil {
		return responseData, err
	}

	return responseData, nil
}
