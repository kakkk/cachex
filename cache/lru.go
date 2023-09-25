package cache

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/kakkk/cachex/internal/model"
	"github.com/kakkk/cachex/internal/utils"
)

type LRUCache[T any] struct {
	cache *expirable.LRU[string, *model.CacheData[T]]
}

// NewLRUCache returns a newly initialize LRUCache implement Cache with ttl and size
func NewLRUCache[T any](size int, ttl time.Duration) *LRUCache[T] {
	return &LRUCache[T]{
		cache: expirable.NewLRU[string, *model.CacheData[T]](size, nil, ttl),
	}
}

func (lc *LRUCache[T]) Get(_ context.Context, key string, expire time.Duration) (T, bool) {
	var zero T
	data, ok := lc.cache.Get(key)
	if !ok {
		return zero, false
	}
	if utils.IsExpired(data.CreateAt, time.Now(), expire) {
		return zero, false
	}
	return data.Data, true
}

func (lc *LRUCache[T]) MGet(ctx context.Context, keys []string, expire time.Duration) map[string]T {
	result := make(map[string]T)
	for _, key := range keys {
		data, ok := lc.Get(ctx, key, expire)
		if !ok {
			continue
		}
		result[key] = data
	}
	return result
}

func (lc *LRUCache[T]) Set(_ context.Context, key string, data T, createTime time.Time) error {
	createAt := utils.ConvertTimestamp(createTime)
	val := utils.NewData(data, createAt)
	lc.cache.Add(key, val)
	return nil
}

func (lc *LRUCache[T]) MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error {
	for k, v := range kvs {
		_ = lc.Set(ctx, k, v, createTime)
	}
	return nil
}

func (lc *LRUCache[T]) Delete(_ context.Context, key string) error {
	lc.cache.Remove(key)
	return nil
}

func (lc *LRUCache[T]) MDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		_ = lc.Delete(ctx, key)
	}
	return nil
}

func (lc *LRUCache[T]) Ping(_ context.Context) (string, error) {
	if lc.cache != nil {
		return "PONG", nil
	}
	return "", errors.New("lru cache not set")
}
