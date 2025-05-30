package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

type Cache struct {
	redisClient  *redis.Client
	memoryCache *cache.Cache
}

func NewCache(redisHost, redisPort string) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create in-memory cache with 5 minute default expiration and 10 minute cleanup interval
	memCache := cache.New(5*time.Minute, 10*time.Minute)

	return &Cache{
		redisClient:  rdb,
		memoryCache: memCache,
	}
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	// Marshal the value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Set in Redis
	err = c.redisClient.Set(context.Background(), key, jsonValue, expiration).Err()
	if err != nil {
		return err
	}

	// Set in memory cache
	c.memoryCache.Set(key, value, expiration)
	return nil
}

func (c *Cache) Get(key string, value interface{}) error {
	// Try memory cache first
	if data, found := c.memoryCache.Get(key); found {
		// Marshal and unmarshal to copy the data into the provided value
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonData, value)
	}

	// Try Redis if not in memory cache
	data, err := c.redisClient.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}

	// Unmarshal the data
	err = json.Unmarshal(data, value)
	if err != nil {
		return err
	}

	// Set in memory cache for future use
	c.memoryCache.Set(key, value, cache.DefaultExpiration)
	return nil
}

func (c *Cache) Delete(key string) error {
	// Delete from Redis
	err := c.redisClient.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}

	// Delete from memory cache
	c.memoryCache.Delete(key)
	return nil
} 