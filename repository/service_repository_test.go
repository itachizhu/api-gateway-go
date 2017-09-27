package repository

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestFindService(t *testing.T)  {
	s := NewServiceRepository()
	service := s.FindService("test-app")
	assert.NotEmpty(t, service)
	assert.Equal(t, int64(1), service.Id)
}