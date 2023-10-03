package cache

import (
	"context"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/internal/utils"
)

func TestNewFreeCache(t *testing.T) {
	ttl := 30 * time.Minute
	fc := NewFreeCache[string](1024*1024, ttl)
	assert.Equal(t, ttl, fc.ttl)
	assert.NotNil(t, fc.cache)
}

func TestFreeCache_Get(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	t.Run("success", func(tt *testing.T) {
		val, _ := utils.MarshalData("success", time.Now().UnixMilli())
		_ = c.Set([]byte("success"), val, 0)
		got, ok := fc.Get(ctx, "success", expire)
		assert.True(tt, ok)
		assert.Equal(tt, "success", got)
	})

	t.Run("expired", func(tt *testing.T) {
		val, _ := utils.MarshalData("expired", time.Now().Add(-25*time.Minute).UnixMilli())
		_ = c.Set([]byte("expired"), val, 0)
		got, ok := fc.Get(ctx, "expired", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)

		got, ok = fc.Get(ctx, "expired", ttl)
		assert.True(tt, ok)
		assert.Equal(tt, "expired", got)
	})

	t.Run("data_default", func(tt *testing.T) {
		val := utils.NewDefaultDataWithMarshal[string](time.Now().UnixMilli())
		_ = c.Set([]byte("data_default"), val, 0)
		got, ok := fc.Get(ctx, "data_default", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("data_nil", func(tt *testing.T) {
		got, ok := fc.Get(ctx, "data_nil", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("unmarshal_error", func(tt *testing.T) {
		_ = c.Set([]byte("unmarshal_error"), []byte("{"), 0)
		got, ok := fc.Get(ctx, "unmarshal_error", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

}

func TestFreeCache_MGet(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	data := map[string]string{
		"multi_success_1": "multi_success_1",
		"multi_success_3": "multi_success_3",
	}

	for k, v := range data {
		val, _ := utils.MarshalData(v, time.Now().UnixMilli())
		_ = c.Set([]byte(k), val, 0)
	}

	keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
	got := fc.MGet(ctx, keys, expire)
	assert.EqualValues(t, data, got)
}

func TestFreeCache_Set(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	t.Run("success", func(tt *testing.T) {
		now := time.Now()
		err := fc.Set(ctx, "success", "success", now)
		assert.Nil(tt, err)
		val, _ := c.Get([]byte("success"))
		want, _ := utils.MarshalData("success", now.UnixMilli())
		assert.JSONEq(tt, string(want), string(val))
	})

	t.Run("free_cache_error", func(tt *testing.T) {
		largeKey := make([]byte, 65536)
		err := fc.Set(ctx, string(largeKey), "success", time.Now())
		assert.NotNil(tt, err)
	})

	t.Run("marshal_error", func(tt *testing.T) {
		type MarshalErrStruct struct {
			A  string      `json:"a"`
			CH chan string `json:"ch"`
		}
		fc2 := &FreeCache[*MarshalErrStruct]{cache: c, ttl: ttl}
		err := fc2.Set(ctx, "marshal_error", &MarshalErrStruct{}, time.Now())
		assert.NotNil(tt, err)
	})
}

func TestFreeCache_MSet(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	t.Run("multi_success", func(tt *testing.T) {
		now := time.Now()
		kvs := map[string]string{
			"multi_success_1": "multi_success_1",
			"multi_success_2": "multi_success_2",
			"multi_success_3": "multi_success_3",
		}
		err := fc.MSet(ctx, kvs, now)
		assert.Nil(tt, err)
		for k, v := range kvs {
			want, _ := utils.MarshalData(v, now.UnixMilli())
			val, _ := c.Get([]byte(k))
			assert.JSONEq(tt, string(want), string(val))
		}
	})

	t.Run("some_error", func(tt *testing.T) {
		now := time.Now()
		largeKey := make([]byte, 65536)
		kvs := map[string]string{
			"some_error_1":   "some_error_1",
			"some_error_2":   "some_error_2",
			string(largeKey): "large_key",
		}
		err := fc.MSet(ctx, kvs, now)
		assert.NotNil(tt, err)
		for _, k := range kvs {
			_, err := c.Get([]byte(k))
			assert.ErrorIs(tt, err, freecache.ErrNotFound)
		}
	})
}

func TestFreeCache_MSetDefault(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	now := time.Now()
	err := fc.SetDefault(ctx, []string{"default_1", "default_2"}, now)
	assert.Nil(t, err)
	want := utils.NewDefaultDataWithMarshal[string](utils.ConvertTimestamp(now))
	got, err := c.Get([]byte("default_1"))
	assert.Nil(t, err)
	assert.JSONEq(t, string(want), string(got))
	got, err = c.Get([]byte("default_2"))
	assert.Nil(t, err)
	assert.JSONEq(t, string(want), string(got))
}

func TestFreeCache_Delete(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	_ = c.Set([]byte("delete"), []byte("delete"), 0)
	val, _ := c.Get([]byte("delete"))
	assert.EqualValues(t, []byte("delete"), val)
	err := fc.Delete(ctx, "delete")
	assert.Nil(t, err)
	_, err = c.Get([]byte("delete"))
	assert.ErrorIs(t, err, freecache.ErrNotFound)
}

func TestFreeCache_MDelete(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}

	kvs := map[string]string{
		"multi_delete_1": "multi_delete_1",
		"multi_delete_2": "multi_delete_2",
		"multi_delete_3": "multi_delete_3",
	}

	for k, v := range kvs {
		_ = c.Set([]byte(k), []byte(v), 0)
		val, _ := c.Get([]byte(k))
		assert.EqualValues(t, v, val)
	}
	err := fc.MDelete(ctx, []string{"multi_delete_1", "multi_delete_2", "multi_delete_3"})
	assert.Nil(t, err)
	for k := range kvs {
		_, err := c.Get([]byte(k))
		assert.ErrorIs(t, err, freecache.ErrNotFound)
	}
}

func TestFreeCache_Ping(t *testing.T) {
	ctx := context.Background()
	c := freecache.NewCache(1024 * 1024)
	ttl := 30 * time.Minute
	fc := &FreeCache[string]{cache: c, ttl: ttl}
	pong, err := fc.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "PONG", pong)

	fc2 := &FreeCache[string]{}
	pong, err = fc2.Ping(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "", pong)
}
