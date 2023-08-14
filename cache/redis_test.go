package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/internal/utils"
)

func TestNewRedisCacheWithClient(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	_ = mr.Set("test", "TestNewRedisCacheWithClient")
	cache := NewRedisCacheWithClient[string](client, ttl)
	assert.Equal(t, ttl, cache.ttl)
	got := cache.client.Get(ctx, "test").Val()
	assert.Equal(t, "TestNewRedisCacheWithClient", got)
}

func TestNewRedisCacheWithOptions(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	_ = mr.Set("test", "TestNewRedisCacheWithOptions")
	cache := NewRedisCacheWithOptions[string](&redis.Options{Addr: mr.Addr()}, ttl)
	assert.Equal(t, ttl, cache.ttl)
	got := cache.client.Get(ctx, "test").Val()
	assert.Equal(t, "TestNewRedisCacheWithOptions", got)
}

func TestRedisCache_Get(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("success", func(tt *testing.T) {
		val, _ := utils.MarshalData("success", time.Now().UnixMilli())
		_ = mr.Set("success", string(val))
		got, ok := rc.Get(ctx, "success", expire)
		assert.True(tt, ok)
		assert.Equal(tt, "success", got)
	})

	t.Run("expired", func(tt *testing.T) {
		val, _ := utils.MarshalData("expired", time.Now().Add(-25*time.Minute).UnixMilli())
		_ = mr.Set("expired", string(val))
		mr.SetTTL("expired", ttl)
		mr.FastForward(25 * time.Minute)

		got, ok := rc.Get(ctx, "expired", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
		assert.True(tt, mr.Exists("expired"))

		got, ok = rc.Get(ctx, "expired", ttl)
		assert.True(tt, ok)
		assert.Equal(tt, "expired", got)

		mr.FastForward(10 * time.Minute)
		got, ok = rc.Get(ctx, "expired", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
		assert.False(tt, mr.Exists("expired"))
	})

	t.Run("data_nil", func(tt *testing.T) {
		got, ok := rc.Get(ctx, "data_nil", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("redis_error", func(tt *testing.T) {
		val, _ := utils.MarshalData("redis_error", time.Now().UnixMilli())
		_ = mr.Set("redis_error", string(val))
		mr.SetError("redis_error")
		defer mr.SetError("")
		got, ok := rc.Get(ctx, "redis_error", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("unmarshal_error", func(tt *testing.T) {
		_ = mr.Set("unmarshal_error", "{")
		got, ok := rc.Get(ctx, "unmarshal_error", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

}

func TestRedisCache_MGet(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("one_success", func(tt *testing.T) {
		val, _ := utils.MarshalData("one_success", time.Now().UnixMilli())
		_ = mr.Set("one_success", string(val))
		got := rc.MGet(ctx, []string{"one_success"}, expire)
		want := map[string]string{
			"one_success": "one_success",
		}
		assert.EqualValues(tt, want, got)
	})

	t.Run("multi_success", func(tt *testing.T) {
		data := map[string]string{
			"multi_success_1": "multi_success_1",
			"multi_success_2": "multi_success_2",
			"multi_success_3": "multi_success_3",
		}
		for k, v := range data {
			val, _ := utils.MarshalData(v, time.Now().UnixMilli())
			_ = mr.Set(k, string(val))
		}
		keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
		got := rc.MGet(ctx, keys, expire)
		assert.EqualValues(tt, data, got)
	})

	t.Run("redis_error", func(tt *testing.T) {
		data := map[string]string{
			"redis_error_1": "redis_error_1",
			"redis_error_2": "redis_error_2",
		}
		for k, v := range data {
			val, _ := utils.MarshalData(v, time.Now().UnixMilli())
			_ = mr.Set(k, string(val))
		}
		mr.SetError("redis_error")
		defer mr.SetError("")
		keys := []string{"redis_error_1", "redis_error_2"}
		got := rc.MGet(ctx, keys, expire)
		assert.EqualValues(tt, map[string]string{}, got)
	})

	t.Run("unmarshal_error", func(tt *testing.T) {

		val, _ := utils.MarshalData("unmarshal_error_1", time.Now().UnixMilli())
		_ = mr.Set("unmarshal_error_1", string(val))
		_ = mr.Set("unmarshal_error_2", "{")

		keys := []string{"unmarshal_error_1", "unmarshal_error_2"}
		want := map[string]string{
			"unmarshal_error_1": "unmarshal_error_1",
		}
		got := rc.MGet(ctx, keys, expire)
		assert.EqualValues(tt, want, got)
	})

	t.Run("some_nil", func(tt *testing.T) {
		data := map[string]string{
			"some_nil_1": "some_nil_1",
			"some_nil_2": "some_nil_2",
		}
		for k, v := range data {
			val, _ := utils.MarshalData(v, time.Now().UnixMilli())
			_ = mr.Set(k, string(val))
		}
		keys := []string{"some_nil_1", "some_nil_2", "some_nil_3"}
		got := rc.MGet(ctx, keys, expire)
		assert.EqualValues(tt, data, got)

	})

	t.Run("some_expired", func(tt *testing.T) {
		data := map[string]string{
			"some_expired_1": "some_expired_1",
			"some_expired_2": "some_expired_2",
			"some_expired_3": "some_expired_3",
		}
		for k, v := range data {
			val, _ := utils.MarshalData(v, time.Now().UnixMilli())
			_ = mr.Set(k, string(val))
		}
		// set 2 expired
		val, _ := utils.MarshalData("some_expired_2", time.Now().Add(-25*time.Minute).UnixMilli())
		_ = mr.Set("some_expired_2", string(val))
		mr.SetTTL("some_expired_2", ttl)
		mr.FastForward(25 * time.Minute)

		keys := []string{"some_expired_1", "some_expired_2", "some_expired_3"}
		want := map[string]string{
			"some_expired_1": "some_expired_1",
			"some_expired_3": "some_expired_3",
		}
		got := rc.MGet(ctx, keys, expire)
		assert.EqualValues(tt, want, got)

		got = rc.MGet(ctx, keys, ttl)
		assert.EqualValues(tt, data, got)

		mr.FastForward(10 * time.Minute)
		got = rc.MGet(ctx, keys, ttl)
		assert.EqualValues(tt, want, got)

	})
}

func TestRedisCache_Set(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("success", func(tt *testing.T) {
		now := time.Now()
		err := rc.Set(ctx, "success", "success", now)
		assert.Nil(tt, err)
		val, _ := mr.Get("success")
		want, _ := utils.MarshalData("success", now.UnixMilli())
		assert.JSONEq(tt, string(want), val)
	})

	t.Run("redis_error", func(tt *testing.T) {
		mr.SetError("unit_test")
		defer mr.SetError("")
		err := rc.Set(ctx, "redis_error", "redis_error", time.Now())
		assert.NotNil(tt, err)
	})

	t.Run("marshal_error", func(tt *testing.T) {
		type MarshalErrStruct struct {
			A  string      `json:"a"`
			CH chan string `json:"ch"`
		}
		rc2 := &RedisCache[*MarshalErrStruct]{client: client, ttl: ttl}
		err := rc2.Set(ctx, "marshal_error", &MarshalErrStruct{}, time.Now())
		assert.NotNil(tt, err)
	})
}

func TestRedisCache_MSet(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("multi_success", func(tt *testing.T) {
		now := time.Now()
		kvs := map[string]string{
			"multi_success_1": "multi_success_1",
			"multi_success_2": "multi_success_2",
			"multi_success_3": "multi_success_3",
		}
		err := rc.MSet(ctx, kvs, now)
		assert.Nil(tt, err)
		for k, v := range kvs {
			want, _ := utils.MarshalData(v, now.UnixMilli())
			val, _ := mr.Get(k)
			assert.JSONEq(tt, string(want), val)
		}
	})

	t.Run("redis_error", func(tt *testing.T) {
		mr.SetError("unit_test")
		defer mr.SetError("")
		kvs := map[string]string{
			"multi_success_1": "multi_success_1",
			"multi_success_2": "multi_success_2",
		}
		err := rc.MSet(ctx, kvs, time.Now())
		assert.NotNil(tt, err)
	})

	t.Run("marshal_error", func(tt *testing.T) {
		type MarshalErrStruct struct {
			A  string      `json:"a"`
			CH chan string `json:"ch"`
		}
		rc2 := &RedisCache[*MarshalErrStruct]{client: client, ttl: ttl}
		kvs := map[string]*MarshalErrStruct{
			"marshal_error": {},
		}
		err := rc2.MSet(ctx, kvs, time.Now())
		assert.NotNil(tt, err)
	})
}

func TestRedisCache_Delete(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("success", func(tt *testing.T) {
		_ = mr.Set("success", "success")
		assert.True(tt, mr.Exists("success"))
		err := rc.Delete(ctx, "success")
		assert.Nil(tt, err)
		assert.False(tt, mr.Exists("success"))
	})

	t.Run("redis_error", func(tt *testing.T) {
		_ = mr.Set("redis_error", "redis_error")
		assert.True(tt, mr.Exists("redis_error"))
		mr.SetError("unit_test")

		err := rc.Delete(ctx, "redis_error")
		assert.NotNil(tt, err)

		mr.SetError("")
		assert.True(tt, mr.Exists("redis_error"))
		err = rc.Delete(ctx, "redis_error")
		assert.Nil(tt, err)
		assert.False(tt, mr.Exists("redis_error"))
	})
}

func TestRedisCache_MDelete(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	t.Run("multi_success", func(tt *testing.T) {
		keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
		for _, key := range keys {
			_ = mr.Set(key, key)
			assert.True(tt, mr.Exists(key))
		}
		err := rc.MDelete(ctx, keys)
		assert.Nil(tt, err)
		for _, key := range keys {
			assert.False(tt, mr.Exists(key))
		}
	})

	t.Run("redis_error", func(tt *testing.T) {
		keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
		for _, key := range keys {
			_ = mr.Set(key, key)
			assert.True(tt, mr.Exists(key))
		}
		mr.SetError("unit_test")
		err := rc.MDelete(ctx, keys)
		assert.NotNil(tt, err)
		mr.SetError("")

		for _, key := range keys {
			assert.True(tt, mr.Exists(key))
		}
		err = rc.MDelete(ctx, keys)
		assert.Nil(tt, err)
		for _, key := range keys {
			assert.False(tt, mr.Exists(key))
		}
	})
}

func TestRedisCache_Ping(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	ttl := 30 * time.Minute
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisCache[string]{client: client, ttl: ttl}

	pong, err := rc.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "PONG", pong)

	mr.SetError("unit_test")
	defer mr.SetError("")
	pong, err = rc.Ping(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "", pong)

	rc2 := &RedisCache[string]{}
	pong, err = rc2.Ping(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "", pong)
}
