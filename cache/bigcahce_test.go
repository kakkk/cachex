package cache

import (
	"context"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/internal/utils"
)

func TestNewBigCache(t *testing.T) {
	got := NewBigCache[string](time.Hour)
	assert.NotNil(t, got)
	assert.NotNil(t, got.cache)
}

func TestNewBigCacheWithConfig(t *testing.T) {
	t.Run("new bigcache error", func(tt *testing.T) {
		cfg := bigcache.Config{
			Shards: 1234,
		}
		got, err := NewBigCacheWithConfig[string](cfg)
		assert.NotNil(tt, err)
		assert.Nil(tt, got)
	})

	t.Run("normal", func(tt *testing.T) {
		got, err := NewBigCacheWithConfig[string](bigcache.DefaultConfig(time.Hour))
		assert.Nil(tt, err)
		assert.NotNil(tt, got)
	})
}

func TestBigCache_Get(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	t.Run("success", func(tt *testing.T) {
		val, _ := utils.MarshalData("success", time.Now().UnixMilli())
		_ = c.Set("success", val)
		got, ok := bc.Get(ctx, "success", expire)
		assert.True(tt, ok)
		assert.Equal(tt, "success", got)
	})

	t.Run("expired", func(tt *testing.T) {
		val, _ := utils.MarshalData("expired", time.Now().Add(-25*time.Minute).UnixMilli())
		_ = c.Set("expired", val)
		got, ok := bc.Get(ctx, "expired", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)

		got, ok = bc.Get(ctx, "expired", ttl)
		assert.True(tt, ok)
		assert.Equal(tt, "expired", got)
	})

	t.Run("data_default", func(tt *testing.T) {
		val := utils.NewDefaultDataWithMarshal[string](time.Now().UnixMilli())
		_ = c.Set("data_default", val)
		got, ok := bc.Get(ctx, "data_default", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("data_nil", func(tt *testing.T) {
		got, ok := bc.Get(ctx, "data_nil", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("unmarshal_error", func(tt *testing.T) {
		_ = c.Set("unmarshal_error", []byte("{"))
		got, ok := bc.Get(ctx, "unmarshal_error", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})
}

func TestBigCache_MGet(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	expire := 20 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	data := map[string]string{
		"multi_success_1": "multi_success_1",
		"multi_success_3": "multi_success_3",
	}

	for k, v := range data {
		val, _ := utils.MarshalData(v, time.Now().UnixMilli())
		_ = c.Set(k, val)
	}

	keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
	got := bc.MGet(ctx, keys, expire)
	assert.EqualValues(t, data, got)
}

func TestBigCache_Set(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	t.Run("success", func(tt *testing.T) {
		now := time.Now()
		err := bc.Set(ctx, "success", "success", now)
		assert.Nil(tt, err)
		val, _ := c.Get("success")
		want, _ := utils.MarshalData("success", now.UnixMilli())
		assert.JSONEq(tt, string(want), string(val))
	})

	t.Run("bigcache_error", func(tt *testing.T) {
		largeVal := make([]byte, 10*1024*1024)
		cc, err := bigcache.New(ctx, bigcache.Config{Shards: 2, HardMaxCacheSize: 64})
		assert.Nil(tt, err)
		bc2 := &BigCache[string]{cache: cc}
		err = bc2.Set(ctx, "bigcache_error", string(largeVal), time.Now())
		assert.NotNil(tt, err)
	})

	t.Run("marshal_error", func(tt *testing.T) {
		type MarshalErrStruct struct {
			A  string      `json:"a"`
			CH chan string `json:"ch"`
		}
		bc2 := &BigCache[*MarshalErrStruct]{cache: c}
		err := bc2.Set(ctx, "marshal_error", &MarshalErrStruct{}, time.Now())
		assert.NotNil(tt, err)
	})
}

func TestBigCache_MSet(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	t.Run("multi_success", func(tt *testing.T) {
		now := time.Now()
		kvs := map[string]string{
			"multi_success_1": "multi_success_1",
			"multi_success_2": "multi_success_2",
			"multi_success_3": "multi_success_3",
		}
		err := bc.MSet(ctx, kvs, now)
		assert.Nil(tt, err)
		for k, v := range kvs {
			want, _ := utils.MarshalData(v, now.UnixMilli())
			val, _ := c.Get(k)
			assert.JSONEq(tt, string(want), string(val))
		}
	})

	t.Run("some_error", func(tt *testing.T) {
		now := time.Now()
		largeVal := make([]byte, 10*1024*1024)
		cc, err := bigcache.New(ctx, bigcache.Config{Shards: 2, HardMaxCacheSize: 64})
		assert.Nil(tt, err)
		bc2 := &BigCache[string]{cache: cc}
		kvs := map[string]string{
			"ok_1":         "ok_1",
			"ok_2":         "ok_2",
			"some_error_2": string(largeVal),
		}
		err = bc2.MSet(ctx, kvs, now)
		assert.NotNil(tt, err)
		for _, k := range kvs {
			_, err := c.Get(k)
			assert.ErrorIs(tt, err, bigcache.ErrEntryNotFound)
		}
	})
}

func TestBigCache_SetDefault(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	now := time.Now()
	err := bc.SetDefault(ctx, []string{"default_1", "default_2"}, now)
	assert.Nil(t, err)
	want := utils.NewDefaultDataWithMarshal[string](utils.ConvertTimestamp(now))
	got, err := c.Get("default_1")
	assert.Nil(t, err)
	assert.JSONEq(t, string(want), string(got))
	got, err = c.Get("default_2")
	assert.Nil(t, err)
	assert.JSONEq(t, string(want), string(got))

}

func TestBigCache_Delete(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	_ = c.Set("delete", []byte("delete"))
	val, _ := c.Get("delete")
	assert.EqualValues(t, []byte("delete"), val)
	err := bc.Delete(ctx, "delete")
	assert.Nil(t, err)
	_, err = c.Get("delete")
	assert.ErrorIs(t, err, bigcache.ErrEntryNotFound)
	err = bc.Delete(ctx, "delete_not_found")
	assert.Nil(t, err)
}

func TestBigCache_MDelete(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}

	kvs := map[string]string{
		"multi_delete_1": "multi_delete_1",
		"multi_delete_2": "multi_delete_2",
		"multi_delete_3": "multi_delete_3",
	}

	for k, v := range kvs {
		_ = c.Set(k, []byte(v))
		val, _ := c.Get(k)
		assert.EqualValues(t, v, val)
	}
	err := bc.MDelete(ctx, []string{"multi_delete_1", "multi_delete_2", "multi_delete_3"})
	assert.Nil(t, err)
	for k := range kvs {
		_, err := c.Get(k)
		assert.ErrorIs(t, err, bigcache.ErrEntryNotFound)
	}
}

func TestBigCache_Ping(t *testing.T) {
	ctx := context.Background()
	ttl := 30 * time.Minute
	c, _ := bigcache.New(ctx, bigcache.DefaultConfig(ttl))
	bc := &BigCache[string]{cache: c}
	pong, err := bc.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "PONG", pong)

	bc2 := &BigCache[string]{}
	pong, err = bc2.Ping(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "", pong)
}
