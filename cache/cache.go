package cache

import "time"

type Cache interface {
	Set(key string, value interface{}, timeout time.Duration) error
	Delete(key ...string) error
	HasKey(key string) bool
	Get(key string) interface{}
	Close() error
}