package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coocood/freecache"

	"github.com/kakkk/cachex/internal/utils"
)

type FreeCache[T any] struct {
	cache *freecache.Cache
	ttl   time.Duration
}

// NewFreeCache returns a newly initialize FreeCache implement Cache by size and ttl
// size: size in bytes, e.g. 100*1024*1024 is 100 MB
// ttl: cache expire ttl, if ttl set 0, cache will not expire
func NewFreeCache[T any](size int, ttl time.Duration) *FreeCache[T] {
	return &FreeCache[T]{
		cache: freecache.NewCache(size),
		ttl:   ttl,
	}
}

func (fc *FreeCache[T]) Get(_ context.Context, key string, expire time.Duration) (T, bool) {
	var zero T
	val, err := fc.cache.Get([]byte(key))
	if err != nil {
		return zero, false
	}
	data, err := utils.UnmarshalData[T](val)
	if err != nil {
		return zero, false
	}
	if utils.IsExpired(data.CreateAt, time.Now(), expire) {
		return zero, false
	}
	return data.Data, true
}

func (fc *FreeCache[T]) MGet(ctx context.Context, keys []string, expire time.Duration) map[string]T {
	result := make(map[string]T)
	for _, key := range keys {
		data, ok := fc.Get(ctx, key, expire)
		if !ok {
			continue
		}
		result[key] = data
	}
	return result
}

func (fc *FreeCache[T]) Set(_ context.Context, key string, data T, createTime time.Time) error {
	createAt := utils.ConvertTimestamp(createTime)
	val, err := utils.MarshalData(data, createAt)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}
	return fc.cache.Set([]byte(key), val, int(fc.ttl.Seconds()))
}

func (fc *FreeCache[T]) MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error {
	success := make([]string, 0, len(kvs))
	for k, v := range kvs {
		err := fc.Set(ctx, k, v, createTime)
		if err != nil {
			_ = fc.MDelete(ctx, success)
			return err
		}
		success = append(success, k)
	}
	return nil
}

func (fc *FreeCache[T]) Delete(_ context.Context, key string) error {
	fc.cache.Del([]byte(key))
	return nil
}

func (fc *FreeCache[T]) MDelete(_ context.Context, keys []string) error {
	for _, key := range keys {
		fc.cache.Del([]byte(key))
	}
	return nil
}

func (fc *FreeCache[T]) Ping(_ context.Context) (string, error) {
	if fc.cache != nil {
		return "PONG", nil
	}
	return "", errors.New("freecache not set")
}
