package cache

// Interface represents simple cache interface
type Interface interface {
	Put(string, interface{}) error
	Get(string, interface{}) error
}

// MeasurableCache represents caches, that has GetMetrics() method
type MeasurableCache interface {
	Interface
	GetMetrics() (put int64, get int64, miss int64, expired int64)
}
