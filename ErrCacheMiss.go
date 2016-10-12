package cache

import "fmt"

// NewErrCacheMiss builds cache MISS error
func NewErrCacheMiss(key string) error {
	return errCacheMiss{key: key}
}

// IsErrCacheMiss returns true if provided error is error,
// returned on cache miss
func IsErrCacheMiss(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(errCacheMiss)
	return ok
}

type errCacheMiss struct {
	key string
}

func (e errCacheMiss) GetKey() string {
	return e.key
}

func (e errCacheMiss) Error() string {
	return fmt.Sprintf(
		"Key \"%s\" not found in cache",
		e.GetKey(),
	)
}
