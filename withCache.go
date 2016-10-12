package cache

import (
	"errors"
	"fmt"
	"reflect"
)

func wrap(src func() (interface{}, error)) func(interface{}) error {
	return func(target interface{}) error {
		if src == nil {
			return errors.New("No source func")
		}
		if reflect.TypeOf(target).Kind() != reflect.Ptr {
			return fmt.Errorf("Pointer receiver expected, but got %T", target)
		}

		data, err := src()
		if err != nil {
			return err
		}
		if data == nil {
			return nil
		}

		if reflect.TypeOf(target).Elem() == reflect.TypeOf(data) {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(data))
		}

		return nil
	}
}

// WithCache function performs cache search and if no entry
// found, runs onMissing function and stores result into cache
func WithCache(cache Interface, name string, target interface{}, onMissing func() (interface{}, error)) error {
	return withCache(cache, name, target, wrap(onMissing))
}

func withCache(cache Interface, name string, target interface{}, onMissing func(interface{}) error) error {
	if cache == nil {
		return onMissing(target)
	}

	// Checking value in cache
	err := cache.Get(name, target)
	if err == nil {
		// Read success
		return nil
	}
	if !IsErrCacheMiss(err) {
		// Got error, that is NOT and cache MISS
		return err
	}

	// Missed cache, generating value
	err = onMissing(target)
	if err != nil {
		// Value generation error
		return err
	}

	// Saving value to cache
	return cache.Put(name, target)
}

// WithDoubleCache performs 2-level cache search with invalidation
func WithDoubleCache(first, second Interface, name string, target interface{}, onMissing func() (interface{}, error)) error {
	return withDoubleCache(first, second, name, target, wrap(onMissing))
}

func withDoubleCache(first, second Interface, name string, target interface{}, onMissing func(interface{}) error) error {
	if first == nil || second == nil {
		return onMissing(target)
	}

	// Checking value in first level cache
	err := first.Get(name, target)
	if err == nil {
		// Read success
		return nil
	}
	if !IsErrCacheMiss(err) {
		// Got error, that is NOT and cache MISS
		return err
	}

	// Checking value in second level cache
	err = second.Get(name, target)
	if err == nil {
		// Read success - copying to first level
		return first.Put(name, target)
	}
	if !IsErrCacheMiss(err) {
		// Got error, that is NOT and cache MISS
		return err
	}

	// Missed both caches
	err = onMissing(target)
	if err != nil {
		// Value generation error
		return err
	}

	// Saving to second level
	err = second.Put(name, target)
	if err == nil {
		// Saving to first level
		err = first.Put(name, target)
	}

	return err
}
