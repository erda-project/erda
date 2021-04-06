package gitmodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	cache := NewMemCache(100, "test")
	testKey := "test"
	testValue := "123456"
	var outValue string
	cache.Set(testKey, testValue)
	err := cache.Get(testKey, &outValue)
	assert.Nil(t, err)
	assert.Equal(t, testValue, outValue)
}
