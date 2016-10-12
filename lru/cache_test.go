package lru

import (
	"github.com/mono83/cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConstructor(t *testing.T) {
	assert := assert.New(t)

	var c cache.MeasurableCache
	var err error

	c, err = New(10, time.Second)
	assert.NotNil(c)
	assert.NoError(err)

	c, err = New(10, time.Duration(0))
	assert.NotNil(c)
	assert.NoError(err)

	c, err = New(1, time.Second)
	assert.Nil(c)
	if assert.Error(err) {
		assert.Equal("At least 2 max items required, but 1 received", err.Error())
	}

	c, err = New(10, time.Millisecond)
	assert.Nil(c)
	if assert.Error(err) {
		assert.Equal("TTL should be zero or at least 10ms, but 1ms received", err.Error())
	}
}

func TestLRUWithoutTTL(t *testing.T) {
	assert := assert.New(t)

	c, err := WithoutTTL(2)
	assert.NoError(err)
	if assert.NotNil(c) {
		assert.NoError(c.Put("key-1", "foo"))
		assert.NoError(c.Put("key-2", "bar"))
		assert.NoError(c.Put("key-3", "baz"))

		var target string
		assert.NoError(c.Get("key-3", &target))
		assert.Equal("baz", target)
		assert.NoError(c.Get("key-2", &target))
		assert.Equal("bar", target)
		assert.True(cache.IsErrCacheMiss(c.Get("key-1", &target)))

		assert.NoError(c.Put("key-4", "var"))
		assert.NoError(c.Get("key-2", &target))
		assert.Equal("bar", target)
		assert.NoError(c.Get("key-4", &target))
		assert.Equal("var", target)
		assert.True(cache.IsErrCacheMiss(c.Get("key-1", &target)))
		assert.True(cache.IsErrCacheMiss(c.Get("key-3", &target)))
	}
}

func TestLRUWithTTL(t *testing.T) {
	assert := assert.New(t)

	c, err := New(10, time.Hour)
	assert.NotNil(c)
	assert.NoError(err)
	assert.NoError(c.Put("key-1", "foo"))
	var target string
	assert.NoError(c.Get("key-1", &target))
	assert.Equal("foo", target)

	c, err = New(10, 10*time.Millisecond)
	assert.NotNil(c)
	assert.NoError(err)
	assert.NoError(c.Put("key-1", "foo"))
	time.Sleep(11 * time.Millisecond)
	err = c.Get("key-1", &target)
	assert.Error(err)
	assert.True(cache.IsErrCacheMiss(err))
}
