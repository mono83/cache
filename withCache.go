package cache

// WithCache function performs cache search and if no entry
// found, runs onMissing function and stores result into cache
func WithCache(cache Interface, name string, target interface{}, onMissing func(interface{}) error) error {
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
