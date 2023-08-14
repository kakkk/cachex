package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kakkk/cachex/internal/utils"
)

type RedisCache[T any] struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCacheWithClient[T any](client *redis.Client, ttl time.Duration) *RedisCache[T] {
	return &RedisCache[T]{
		client: client,
		ttl:    ttl,
	}
}

func NewRedisCacheWithOptions[T any](options *redis.Options, ttl time.Duration) *RedisCache[T] {
	return &RedisCache[T]{
		client: redis.NewClient(options),
		ttl:    ttl,
	}
}

func (rc *RedisCache[T]) Get(ctx context.Context, key string, expire time.Duration) (T, bool) {
	var zero T
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return zero, false
	}
	data, err := utils.UnmarshalData[T]([]byte(val))
	if err != nil {
		return zero, false
	}
	if utils.IsExpired(data.CreateAt, time.Now(), expire) {
		return zero, false
	}
	return data.Data, true
}

func (rc *RedisCache[T]) MGet(ctx context.Context, keys []string, expire time.Duration) map[string]T {
	now := time.Now()
	values, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return make(map[string]T)
	}
	result := make(map[string]T, len(keys))
	for i, key := range keys {
		val, ok := values[i].(string)
		if !ok {
			continue
		}
		data, err := utils.UnmarshalData[T]([]byte(val))
		if err != nil {
			continue
		}
		if utils.IsExpired(data.CreateAt, now, expire) {
			continue
		}
		result[key] = data.Data
	}
	return result
}

func (rc *RedisCache[T]) Set(ctx context.Context, key string, data T, createTime time.Time) error {
	createAt := utils.ConvertTimestamp(createTime)
	val, err := utils.MarshalData(data, createAt)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}
	return rc.client.Set(ctx, key, val, rc.ttl).Err()
}

func (rc *RedisCache[T]) MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error {
	pipe := rc.client.Pipeline()
	createAt := utils.ConvertTimestamp(createTime)
	for k, v := range kvs {
		val, err := utils.MarshalData(v, createAt)
		if err != nil {
			return fmt.Errorf("marshal error: %v", err)
		}
		pipe.Set(ctx, k, val, rc.ttl)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache[T]) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisCache[T]) MDelete(ctx context.Context, keys []string) error {
	return rc.client.Del(ctx, keys...).Err()
}

func (rc *RedisCache[T]) Ping(ctx context.Context) (string, error) {
	if rc.client == nil {
		return "", errors.New("redis client not set")
	}
	return rc.client.Ping(ctx).Result()
}
