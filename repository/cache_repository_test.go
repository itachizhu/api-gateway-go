package repository

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"log"
)

func TestFindCache(t *testing.T) {
	cache := new(CacheRepository).FindCache(int64(1), []string{"/", "/hello"})
	assert.NotEmpty(t, cache)
	log.Printf("====path: %v", cache.Path)
	assert.Equal(t, "/hello", cache.Path)
}