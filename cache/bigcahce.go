package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/kakkk/cachex/internal/utils"
)

type BigCache[T any] struct {
	cache *bigcache.BigCache
}

// NewBigCache returns a newly initialize BigCache implement Cache with ttl,
// the bigcache config use bigcache.DefaultConfig
func NewBigCache[T any](ttl time.Duration) *BigCache[T] {
	c, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(ttl))
	return &BigCache[T]{
		cache: c,
	}
}

// NewBigCacheWithConfig returns a newly initialize BigCache implement Cache with bigcache config,
//
// cfg: bigcache.Config
func NewBigCacheWithConfig[T any](cfg bigcache.Config) (*BigCache[T], error) {
	c, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &BigCache[T]{
		cache: c,
	}, nil
}

func (bc *BigCache[T]) Get(_ context.Context, key string, expire time.Duration) (T, bool) {
	var zero T
	val, err := bc.cache.Get(key)
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

func (bc *BigCache[T]) MGet(ctx context.Context, keys []string, expire time.Duration) map[string]T {
	result := make(map[string]T)
	for _, key := range keys {
		data, ok := bc.Get(ctx, key, expire)
		if !ok {
			continue
		}
		result[key] = data
	}
	return result
}

func (bc *BigCache[T]) Set(_ context.Context, key string, data T, createTime time.Time) error {
	createAt := utils.ConvertTimestamp(createTime)
	val, err := utils.MarshalData(data, createAt)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}
	return bc.cache.Set(key, val)
}

func (bc *BigCache[T]) MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error {
	success := make([]string, 0, len(kvs))
	for k, v := range kvs {
		err := bc.Set(ctx, k, v, createTime)
		if err != nil {
			_ = bc.MDelete(ctx, success)
			return err
		}
		success = append(success, k)
	}
	return nil
}

func (bc *BigCache[T]) Delete(_ context.Context, key string) error {
	err := bc.cache.Delete(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (bc *BigCache[T]) MDelete(ctx context.Context, keys []string) error {
	var errs []error
	for _, key := range keys {
		err := bc.Delete(ctx, key)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (bc *BigCache[T]) Ping(_ context.Context) (string, error) {
	if bc.cache != nil {
		return "PONG", nil
	}
	return "", errors.New("bigcache not set")
}
