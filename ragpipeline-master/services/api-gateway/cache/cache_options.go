package cache

import "time"

// CacheOption configures a cache instance.
type CacheOption interface {
	apply(*cacheOptions)
}

type cacheOptions struct {
	ttl     time.Duration
	maxSize int
}

type ttlOption struct {
	TTL time.Duration
}

func (o ttlOption) apply(opts *cacheOptions) { opts.ttl = o.TTL }

type maxSizeOption struct {
	Size int
}

func (o maxSizeOption) apply(opts *cacheOptions) { opts.maxSize = o.Size }

// WithCacheTTL sets the cache time-to-live duration.
func WithCacheTTL(ttl time.Duration) CacheOption {
	return ttlOption{TTL: ttl}
}

// WithCacheMaxSize sets the maximum number of entries in the cache.
func WithCacheMaxSize(size int) CacheOption {
	return maxSizeOption{Size: size}
}
