package cache

import (
	"context"
	"time"
)

type Cache[T any] interface {
	Get(ctx context.Context, key string, expire time.Duration) (data T, ok bool)
	MGet(ctx context.Context, keys []string, expire time.Duration) (data map[string]T)
	Set(ctx context.Context, key string, data T, createTime time.Time) error
	MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error
	SetDefault(ctx context.Context, keys []string, createTime time.Time) error
	Delete(ctx context.Context, key string) error
	MDelete(ctx context.Context, keys []string) error
	Ping(ctx context.Context) (string, error)
}
