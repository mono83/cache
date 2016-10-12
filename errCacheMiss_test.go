package cache

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrCacheMiss(t *testing.T) {
	assert := assert.New(t)

	assert.Error(NewErrCacheMiss("foo"))
	assert.Equal("Key \"foo\" not found in cache", NewErrCacheMiss("foo").Error())
	assert.True(IsErrCacheMiss(NewErrCacheMiss("foo")))
	assert.False(IsErrCacheMiss(errors.New("Key \"foo\" not found in cache")))
	assert.False(IsErrCacheMiss(nil))
}
