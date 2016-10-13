package lru

import (
	"fmt"
	"github.com/golang/groupcache/lru"
	"github.com/mono83/cache"
	"reflect"
	"sync"
	"time"
)

type cacheNode struct {
	expiryAt  time.Time
	valueType reflect.Type
	value     interface{}
}

func (c cacheNode) into(target interface{}) error {
	receivedType := reflect.TypeOf(target).Elem()

	if c.valueType == receivedType {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(c.value))
		return nil
	}
	if c.valueType.AssignableTo(receivedType) {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(c.value))
		return nil
	}

	return fmt.Errorf("Unable to write %s into %s", c.valueType.String(), receivedType.String())
}

func (c cacheNode) alive() bool {
	return c.expiryAt.IsZero() || time.Now().Before(c.expiryAt)
}

// cacher is LRU implementation of cache.Interface
type cacher struct {
	l   *lru.Cache
	m   sync.Mutex
	ttl time.Duration

	statGet, statGetMiss, statExpire, statPut int64
}

// New builds and returns new LRU cache instance with TTL
func New(max int, ttl time.Duration) (cache.MeasurableCache, error) {
	if max < 2 {
		return nil, fmt.Errorf("At least 2 max items required, but %d received", max)
	}
	if ttl.Nanoseconds() != 0 && ttl.Seconds() < 0.01 {
		return nil, fmt.Errorf("TTL should be zero or at least 10ms, but %s received", ttl.String())
	}
	lc := new(cacher)
	lc.l = lru.New(max)
	lc.ttl = ttl

	return lc, nil
}

// WithoutTTL returns LRU cache without TTL
func WithoutTTL(max int) (cache.MeasurableCache, error) {
	return New(max, time.Duration(0))
}

// Get writes value from cache into {value} variable
// Will return ErrCacheMiss if no entry found
func (c *cacher) Get(key string, value interface{}) error {
	// Checking pointer type
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return fmt.Errorf("Target value must be pointer, but %T received", value)
	}

	c.m.Lock()
	c.statGet++
	v, ok := c.l.Get(lru.Key(key))
	defer c.m.Unlock()
	if !ok {
		c.statGetMiss++
		return cache.NewErrCacheMiss(key)
	}

	cv, _ := v.(cacheNode)
	if !cv.alive() {
		c.statExpire++
		c.l.Remove(lru.Key(key))

		return cache.NewErrCacheMiss(key)
	}

	return cv.into(value)
}

// Put stores value into cache
func (c *cacher) Put(key string, value interface{}) error {
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		// Copying from pointer
		value = reflect.ValueOf(value).Elem().Interface()
		valueType = valueType.Elem()
	}

	node := cacheNode{value: value, valueType: valueType}
	if c.ttl.Nanoseconds() != 0 {
		node.expiryAt = time.Now().Add(c.ttl)
	}

	c.m.Lock()
	defer c.m.Unlock()
	c.l.Add(lru.Key(key), node)
	c.statPut++
	return nil
}

func (c *cacher) GetMetrics() (put int64, get int64, miss int64, expired int64) {
	put = c.statPut
	get = c.statGet
	miss = c.statGetMiss
	expired = c.statGetMiss

	return
}
