package cache

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/internal/model"
	"github.com/kakkk/cachex/internal/utils"
)

func TestNewLRUCache(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	lruCache := NewLRUCache[string](size, ttl)
	assert.NotNil(t, lruCache)
}

func TestLRUCache_Get(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	expire := 20 * time.Minute
	lc := &LRUCache[string]{cache: c}

	t.Run("success", func(tt *testing.T) {
		val := utils.NewData("success", time.Now().UnixMilli())
		_ = c.Add("success", val)
		got, ok := lc.Get(ctx, "success", expire)
		assert.True(tt, ok)
		assert.Equal(tt, "success", got)
	})

	t.Run("expired", func(tt *testing.T) {
		val := utils.NewData("expired", time.Now().Add(-25*time.Minute).UnixMilli())
		_ = c.Add("expired", val)
		got, ok := lc.Get(ctx, "expired", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)

		got, ok = lc.Get(ctx, "expired", ttl)
		assert.True(tt, ok)
		assert.Equal(tt, "expired", got)
	})

	t.Run("data_default", func(tt *testing.T) {
		val := utils.NewDefaultData[string](time.Now().UnixMilli())
		_ = c.Add("data_default", val)
		got, ok := lc.Get(ctx, "data_default", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("data_nil", func(tt *testing.T) {
		got, ok := lc.Get(ctx, "data_nil", expire)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

}

func TestLRUCache_MGet(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	expire := 20 * time.Minute
	lc := &LRUCache[string]{cache: c}

	data := map[string]string{
		"multi_success_1": "multi_success_1",
		"multi_success_3": "multi_success_3",
	}

	for k, v := range data {
		val := utils.NewData(v, time.Now().UnixMilli())
		_ = c.Add(k, val)
	}

	keys := []string{"multi_success_1", "multi_success_2", "multi_success_3"}
	got := lc.MGet(ctx, keys, expire)
	assert.EqualValues(t, data, got)
}

func TestLRUCache_Set(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}

	now := time.Now()
	err := lc.Set(ctx, "success", "success", now)
	assert.Nil(t, err)
	val, _ := c.Get("success")
	want := utils.NewData("success", now.UnixMilli())
	assert.EqualValues(t, want, val)
}

func TestLRUCache_MSet(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}

	now := time.Now()
	kvs := map[string]string{
		"multi_success_1": "multi_success_1",
		"multi_success_2": "multi_success_2",
		"multi_success_3": "multi_success_3",
	}
	err := lc.MSet(ctx, kvs, now)
	assert.Nil(t, err)
	for k, v := range kvs {
		want := utils.NewData(v, now.UnixMilli())
		val, _ := c.Get(k)
		assert.EqualValues(t, want, val)
	}
}

func TestLRUCache_SetDefault(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}

	now := time.Now()
	err := lc.SetDefault(ctx, []string{"default_1", "default_2"}, now)
	assert.Nil(t, err)
	want := utils.NewDefaultData[string](utils.ConvertTimestamp(now))
	got, _ := c.Get("default_1")
	assert.EqualValues(t, want, got)
	got, _ = c.Get("default_2")
	assert.EqualValues(t, want, got)
}

func TestLRUCache_Delete(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}

	data := utils.NewData("delete", time.Now().UnixMilli())
	_ = c.Add("delete", data)
	val, _ := c.Get("delete")
	assert.EqualValues(t, data, val)
	err := lc.Delete(ctx, "delete")
	assert.Nil(t, err)
	_, ok := c.Get("delete")
	assert.False(t, ok)
}

func TestLRUCache_MDelete(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}

	kvs := map[string]string{
		"multi_delete_1": "multi_delete_1",
		"multi_delete_2": "multi_delete_2",
		"multi_delete_3": "multi_delete_3",
	}

	for k, v := range kvs {
		data := utils.NewData(v, time.Now().UnixMilli())
		c.Add(k, data)
		val, _ := c.Get(k)
		assert.EqualValues(t, data, val)
	}
	err := lc.MDelete(ctx, []string{"multi_delete_1", "multi_delete_2", "multi_delete_3"})
	assert.Nil(t, err)
	for k := range kvs {
		_, ok := c.Get(k)
		assert.False(t, ok)
	}
}

func TestLRUCache_Ping(t *testing.T) {
	ttl := time.Minute * 30
	size := 10
	ctx := context.Background()
	c := expirable.NewLRU[string, *model.CacheData[string]](size, nil, ttl)
	lc := &LRUCache[string]{cache: c}
	pong, err := lc.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "PONG", pong)

	lc2 := &LRUCache[string]{}
	pong, err = lc2.Ping(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "", pong)
}
