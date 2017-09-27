package domain

import (
	"github.com/itachizhu/api-gateway-go/cache/redis"
	"github.com/itachizhu/api-gateway-go/cache"
	redigo "github.com/garyburd/redigo/redis"
)

var redisCache cache.Cache

func init()  {
	redisCache = redis.NewRedisCache()
}

func Close()  {
	redisCache.Close()
}

func get(key string) string  {
	value := redisCache.Get(key)
	if value != nil {
		val, err := redigo.String(value, nil)
		if err != nil {
			return ""
		}
		return val
	}
	return ""
}