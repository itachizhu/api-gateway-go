package redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/itachizhu/api-gateway-go/cache"
	"time"
)

var pool *redis.Pool

type Cache struct {

}

func NewRedisCache() cache.Cache {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", "localhost:6379", redis.DialConnectTimeout(20 * time.Second), redis.DialReadTimeout(10 * time.Second), redis.DialWriteTimeout(10 * time.Second))
		if err != nil {
			return nil, err
		}
		_, selecter := c.Do("SELECT", 0)
		if selecter != nil {
			c.Close()
			return nil, selecter
		}
		return
	}

	pool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   10,
		IdleTimeout: 180 * time.Second,
		Wait:        true,
		Dial:        dialFunc,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return &Cache{}
}

func (r *Cache) Set(key string, value interface{}, timeout time.Duration) error {
	client := pool.Get()
	defer client.Close()
	if timeout > 0 {
		if _, err := client.Do("SET", key, value, "EX", int64(timeout)); err != nil {
			return err
		}
	} else {
		if _, err := client.Do("SET", key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *Cache) Delete(keys ...string) error {
	client := pool.Get()
	defer client.Close()
	for _, key := range keys {
		_, err := client.Do("DEL", key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Cache) HasKey(key string) bool {
	client := pool.Get()
	defer client.Close()
	_, err := redis.Bool(client.Do("EXISTS", key))
	if err == nil {
		return true
	}
	return false
}

func (r *Cache) Get(key string) interface{} {
	client := pool.Get()
	defer client.Close()
	value, err := client.Do("GET", key)
	if err == nil {
		return value
	}
	return nil
}

func (r *Cache) Close() error {
	if pool != nil {
		err := pool.Close()
		pool = nil
		return err
	}
	return nil
}
