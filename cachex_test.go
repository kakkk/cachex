package cachex

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/cache"
	"github.com/kakkk/cachex/internal/consts"
	"github.com/kakkk/cachex/internal/logger"
)

func TestCacheX_Set(t *testing.T) {
	key, value := "key", "value"
	t.Run("all success", func(tt *testing.T) {
		setCount := 0
		set := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return nil
		}
		cache0 := cache.NewCacheMocker[string]().MockSet(set)
		cache1 := cache.NewCacheMocker[string]().MockSet(set)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Set(context.Background(), key, value)
		assert.Nil(tt, err)
		assert.Equal(tt, 2, setCount)
	})

	t.Run("cache0 success cache1 fail", func(tt *testing.T) {
		setCount := 0
		set0 := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return nil
		}
		set1 := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockSet(set0)
		cache1 := cache.NewCacheMocker[string]().MockSet(set1)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Set(context.Background(), key, value)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.False(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("cache0 fail cache1 success", func(tt *testing.T) {
		setCount := 0
		set0 := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		set1 := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return nil
		}
		cache0 := cache.NewCacheMocker[string]().MockSet(set0)
		cache1 := cache.NewCacheMocker[string]().MockSet(set1)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Set(context.Background(), key, value)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.False(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("all fail", func(tt *testing.T) {
		setCount := 0
		set := func(ctx context.Context, key string, data string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockSet(set)
		cache1 := cache.NewCacheMocker[string]().MockSet(set)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Set(context.Background(), key, value)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("panic recover", func(tt *testing.T) {
		set := func(ctx context.Context, key string, data string, createTime time.Time) error {
			panic("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockSet(set)
		cache1 := cache.NewCacheMocker[string]().MockSet(set)
		cx := &CacheX[string, string]{
			logger:     logger.NewDefaultLogger(),
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Set(context.Background(), key, value)
		assert.NotNil(tt, err)
	})
}

func TestCacheX_MSet(t *testing.T) {
	kvs := map[string]string{
		"key_1": "value_1",
		"key_2": "value_2",
		"key_3": "value_3",
	}
	t.Run("all success", func(tt *testing.T) {
		setCount := 0
		mSet := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return nil
		}
		cache0 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cache1 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}

		err := cx.MSet(context.Background(), kvs)
		assert.Nil(tt, err)
		assert.Equal(tt, 2, setCount)
	})

	t.Run("kvs is nil", func(tt *testing.T) {
		cx := &CacheX[string, string]{}
		err := cx.MSet(context.Background(), nil)
		assert.Nil(tt, err)
	})

	t.Run("cache0 success cache1 fail", func(tt *testing.T) {
		setCount := 0
		mSet0 := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return nil
		}
		mSet1 := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockMSet(mSet0)
		cache1 := cache.NewCacheMocker[string]().MockMSet(mSet1)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}

		err := cx.MSet(context.Background(), kvs)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.False(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("cache0 fail cache1 success", func(tt *testing.T) {
		setCount := 0
		mSet0 := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		mSet1 := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return nil
		}
		cache0 := cache.NewCacheMocker[string]().MockMSet(mSet0)
		cache1 := cache.NewCacheMocker[string]().MockMSet(mSet1)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}

		err := cx.MSet(context.Background(), kvs)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.False(tt, cErr.GetErrorLevels()[1])
		assert.True(tt, cErr.GetErrorLevels()[0])
	})

	t.Run("all fail", func(tt *testing.T) {
		setCount := 0
		mSet := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			setCount++
			return errors.New("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cache1 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}

		err := cx.MSet(context.Background(), kvs)
		assert.NotNil(tt, err)
		assert.Equal(tt, 2, setCount)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("panic recover", func(tt *testing.T) {
		mSet := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			panic("unit_test")
		}
		cache0 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cache1 := cache.NewCacheMocker[string]().MockMSet(mSet)
		cx := &CacheX[string, string]{
			logger:     logger.NewDefaultLogger(),
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}

		err := cx.MSet(context.Background(), kvs)
		assert.NotNil(tt, err)
	})
}

func TestCacheX_Get(t *testing.T) {
	key, value := "key", "value"

	t.Run("cache hit 1", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return value, true
			})
		cache1 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return value, true
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			hitCallback: func(name string, level int) {
				assert.Equal(tt, 1, level)
			},
		}
		got, ok := cx.Get(context.Background(), key, 0)
		assert.True(tt, ok)
		assert.Equal(tt, value, got)
	})

	t.Run("cache hit 0", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return value, true
			})
		cache1 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return "", false
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			hitCallback: func(name string, level int) {
				assert.Equal(tt, 0, level)
			},
		}
		got, ok := cx.Get(context.Background(), key, 0)
		assert.True(tt, ok)
		assert.Equal(tt, value, got)
	})

	t.Run("cache not hit", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return "", false
			})
		cache1 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return "", false
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			hitCallback: func(name string, level int) {
				assert.Equal(tt, consts.CacheLevelSource, level)
			},
		}
		got, ok := cx.Get(context.Background(), key, 0)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("panic recover", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				panic("unit_test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			logger:     logger.NewDefaultLogger(),
		}
		got, ok := cx.Get(context.Background(), key, 0)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})
}

func TestCacheX_MGet(t *testing.T) {
	kvs := map[string]string{
		"key_1": "value_1",
		"key_2": "value_2",
		"key_3": "value_3",
	}

	t.Run("cache hit 1", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{}
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return kvs
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			mHitCallback: func(name string, level int, times int) {
				assert.Equal(tt, 1, level)
				assert.Equal(tt, 3, times)
			},
		}
		got := cx.MGet(context.Background(), []string{"key_1", "key_2", "key_3"}, 0)
		assert.EqualValues(tt, kvs, got)
	})

	t.Run("cache hit 0", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return kvs
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{}
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			mHitCallback: func(name string, level int, times int) {
				assert.Equal(tt, 0, level)
				assert.Equal(tt, 3, times)
			},
		}
		got := cx.MGet(context.Background(), []string{"key_1", "key_2", "key_3"}, 0)
		assert.EqualValues(tt, kvs, got)
	})

	t.Run("cache hit 1 and 0", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{
					"key_1": "value_1",
					"key_3": "value_3",
				}
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{
					"key_2": "value_2",
				}
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			mHitCallback: func(name string, level int, times int) {
				if level == 0 {
					assert.Equal(tt, 2, times)
				} else if level == 1 {
					assert.Equal(tt, 1, times)

				}
			},
		}
		got := cx.MGet(context.Background(), []string{"key_1", "key_2", "key_3"}, 0)
		assert.EqualValues(tt, kvs, got)
	})

	t.Run("some not hit or from source", func(tt *testing.T) {
		mSet := func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
			assert.EqualValues(t, map[string]string{"key_2": "value_2"}, kvs)
			return nil
		}
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{}
			}).
			MockMSet(mSet)
		cache1 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{
					"key_1": "value_1",
				}
			}).
			MockMSet(mSet)
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			mHitCallback: func(name string, level int, times int) {
				if level == 1 {
					assert.Equal(tt, 1, times)
				} else if level == -1 {
					assert.Equal(tt, 2, times)
				}
			},
			mGetRealData: func(ctx context.Context, keys []string) (data map[string]string, err error) {
				return map[string]string{"key_2": "value_2"}, err
			},
		}
		got := cx.MGet(context.Background(), []string{"key_1", "key_2", "key_3"}, 0)
		want := map[string]string{
			"key_1": "value_1",
			"key_2": "value_2",
		}
		assert.EqualValues(tt, want, got)
	})

	t.Run("panic recover", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				panic("unit_test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			logger:     logger.NewDefaultLogger(),
		}
		got := cx.MGet(context.Background(), []string{"key_1", "key_2", "key_3"}, 0)
		assert.NotNil(tt, got)
		assert.EqualValues(tt, map[string]string{}, got)
	})
}

func TestCacheX_Delete(t *testing.T) {
	key := "key"

	t.Run("success", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return nil
			})
		cache1 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Delete(context.Background(), key)
		assert.Nil(tt, err)
	})

	t.Run("cache0 success cache1 fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return nil
			})
		cache1 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return errors.New("test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Delete(context.Background(), key)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.False(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("cache1 success cache0 fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return errors.New("test")
			})
		cache1 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Delete(context.Background(), key)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.False(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("all fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return errors.New("test")
			})
		cache1 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				assert.Equal(tt, key, k)
				return errors.New("test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.Delete(context.Background(), key)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("panic recover", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockDelete(func(ctx context.Context, k string) error {
				panic("unit_test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			logger:     logger.NewDefaultLogger(),
		}
		err := cx.Delete(context.Background(), key)
		assert.NotNil(tt, err)
		var cErr CacheError
		assert.False(tt, errors.As(err, &cErr))
	})
}

func TestCacheX_MDelete(t *testing.T) {
	keys := []string{"key_1", "key_2"}

	t.Run("success", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return nil
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.MDelete(context.Background(), keys)
		assert.Nil(tt, err)
	})

	t.Run("cache0 success cache1 fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return nil
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return errors.New("test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.MDelete(context.Background(), keys)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.False(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("cache1 success cache0 fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return errors.New("test")
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.MDelete(context.Background(), keys)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.False(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("all fail", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return errors.New("test")
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				assert.Equal(tt, keys, k)
				return errors.New("test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
		}
		err := cx.MDelete(context.Background(), keys)
		assert.NotNil(tt, err)
		var cErr CacheError
		errors.As(err, &cErr)
		assert.True(tt, cErr.GetErrorLevels()[0])
		assert.True(tt, cErr.GetErrorLevels()[1])
	})

	t.Run("panic recover", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMDelete(func(ctx context.Context, k []string) error {
				panic("unit_test")
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			logger:     logger.NewDefaultLogger(),
		}
		err := cx.MDelete(context.Background(), keys)
		assert.NotNil(tt, err)
		var cErr CacheError
		assert.False(tt, errors.As(err, &cErr))
	})
}

func TestCacheX_getRealDataInternal(t *testing.T) {
	ctx := context.Background()

	t.Run("get real data not set", func(tt *testing.T) {
		cx := &CacheX[string, string]{}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("get real data success", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockSet(func(ctx context.Context, key string, data string, createTime time.Time) error {
				assert.Equal(tt, "k", key)
				assert.Equal(tt, "v", data)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				return "v", nil
			},
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.True(tt, ok)
		assert.Equal(tt, "v", got)
	})

	t.Run("get real data fail and not allow downgrade", func(tt *testing.T) {
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				return "", errors.New("test")
			},
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("allow downgrade but data not found", func(tt *testing.T) {
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				return "", ErrNotFound
			},
			allowDowngrade: true,
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})

	t.Run("downgrade got data", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				assert.Equal(tt, "k", key)
				if expire == time.Hour {
					return "v", true
				}
				return "", false
			})
		testErr := errors.New("test")
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				return "", testErr
			},
			allowDowngrade:           true,
			downgradeCacheExpireTime: time.Hour,
			downgradeCallback: func(ctx context.Context, key string, err error) {
				assert.Equal(tt, "k", key)
				assert.ErrorIs(tt, err, testErr)
			},
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.True(tt, ok)
		assert.Equal(tt, "v", got)
	})

	t.Run("panic downgrade got data", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				assert.Equal(tt, "k", key)
				if expire == time.Hour {
					return "v", true
				}
				return "", false
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				panic("test")
			},
			logger:                   logger.NewDefaultLogger(),
			allowDowngrade:           true,
			downgradeCacheExpireTime: time.Hour,
			downgradeCallback: func(ctx context.Context, key string, err error) {
				assert.Equal(tt, "k", key)
				assert.Contains(tt, err.Error(), "[panic recover]")
			},
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.True(tt, ok)
		assert.Equal(tt, "v", got)
	})

	t.Run("downgrade but data not found", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				assert.Equal(tt, "k", key)
				return "", false
			})
		testErr := errors.New("test")
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			getRealData: func(ctx context.Context, key string) (data string, err error) {
				return "", testErr
			},
			allowDowngrade:           true,
			downgradeCacheExpireTime: time.Hour,
			downgradeCallback: func(ctx context.Context, key string, err error) {
				assert.Equal(tt, "k", key)
				assert.ErrorIs(tt, err, testErr)
			},
		}
		got, ok := cx.getRealDataInternal(ctx, "k")
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
	})
}

func TestCacheX_mGetRealDataInternal(t *testing.T) {
	ctx := context.Background()
	keys := []string{"k_1", "k_2", "k_3"}

	t.Run("mget real data not set", func(tt *testing.T) {
		cx := &CacheX[string, string]{}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.Empty(tt, got)
	})

	t.Run("mget real data success", func(tt *testing.T) {
		data := map[string]string{
			"k_1": "v_1",
			"k_2": "v_2",
			"k_3": "v_3",
		}
		cache0 := cache.NewCacheMocker[string]().
			MockMSet(func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
				want := map[string]string{
					"k_1": "v_1",
					"k_2": "v_2",
					"k_3": "v_3",
				}
				assert.EqualValues(tt, want, kvs)
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			mGetRealData: func(ctx context.Context, keys []string) (map[string]string, error) {
				return data, nil
			},
		}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.EqualValues(tt, data, got)
	})

	t.Run("mget real data fail and not allow downgrade", func(tt *testing.T) {
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			mGetRealData: func(ctx context.Context, keys []string) (map[string]string, error) {
				return nil, errors.New("test")
			},
		}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.Empty(tt, got)
	})

	t.Run("allow downgrade but data not found", func(tt *testing.T) {
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			mGetRealData: func(ctx context.Context, keys []string) (map[string]string, error) {
				return nil, ErrNotFound
			},
			allowDowngrade: true,
		}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.Empty(tt, got)
	})

	t.Run("downgrade got data", func(tt *testing.T) {
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				if expire == time.Hour {
					data := map[string]string{
						"k_1": "v_1",
						"k_2": "v_2",
					}
					return data
				}
				return map[string]string{}
			})
		cache1 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				if expire == time.Hour {
					data := map[string]string{
						"k_3": "v_3",
					}
					return data
				}
				return map[string]string{}
			})
		testErr := errors.New("test")
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0, cache1},
			mGetRealData: func(ctx context.Context, keys []string) (map[string]string, error) {
				return nil, testErr
			},
			allowDowngrade:           true,
			downgradeCacheExpireTime: time.Hour,
			mDowngradeCallback: func(ctx context.Context, k []string, err error) {
				assert.EqualValues(tt, keys, k)
				assert.ErrorIs(tt, err, testErr)
			},
		}
		want := map[string]string{
			"k_1": "v_1",
			"k_2": "v_2",
			"k_3": "v_3",
		}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.EqualValues(tt, want, got)
	})

	t.Run("panic downgrade got data", func(tt *testing.T) {
		want := map[string]string{
			"k_1": "v_1",
			"k_2": "v_2",
			"k_3": "v_3",
		}
		cache0 := cache.NewCacheMocker[string]().
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				if expire == time.Hour {
					return map[string]string{
						"k_1": "v_1",
						"k_2": "v_2",
						"k_3": "v_3",
					}
				}
				return nil
			})
		cx := &CacheX[string, string]{
			getDataKey: func(key string) string { return key },
			caches:     []cache.Cache[string]{cache0},
			mGetRealData: func(ctx context.Context, keys []string) (data map[string]string, err error) {
				panic("test")
			},
			logger:                   logger.NewDefaultLogger(),
			allowDowngrade:           true,
			downgradeCacheExpireTime: time.Hour,
			mDowngradeCallback: func(ctx context.Context, k []string, err error) {
				assert.EqualValues(tt, keys, k)
				assert.Contains(tt, err.Error(), "[panic recover]")
			},
		}
		got := cx.mGetRealDataInternal(ctx, keys)
		assert.EqualValues(tt, want, got)
	})

}

func TestCacheX_mGetDataKeys(t *testing.T) {
	keys := []string{"k_1", "k_2", "k_3"}
	cx := CacheX[string, string]{
		getDataKey: func(key string) string {
			return fmt.Sprintf("k_%s", key)
		},
	}
	want := []string{"k_k_1", "k_k_2", "k_k_3"}
	got := cx.mGetDataKeys(keys)
	assert.EqualValues(t, want, got)
}

func TestCacheX_hit(t *testing.T) {
	ctx := context.Background()
	cx := CacheX[string, string]{
		logger: logger.NewDefaultLogger(),
	}
	cx.hit(ctx, 1)
	cx.hitCallback = func(name string, level int) {
		assert.Equal(t, 1, level)
	}
	cx.hit(ctx, 1)
}

func TestCacheX_mHit(t *testing.T) {
	ctx := context.Background()
	cx := CacheX[string, string]{
		logger: logger.NewDefaultLogger(),
	}
	cx.mHit(ctx, 1, 3)
	cx.mHitCallback = func(name string, level int, times int) {
		assert.Equal(t, 1, level)
		assert.Equal(t, 3, times)
	}
	cx.mHit(ctx, 1, 3)
}

func TestCacheX_downgrade(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test")
	cx := CacheX[string, string]{
		logger: logger.NewDefaultLogger(),
	}
	cx.downgrade(ctx, "k", testErr)
	cx.downgradeCallback = func(ctx context.Context, key string, err error) {
		assert.Equal(t, "k", key)
		assert.ErrorIs(t, err, testErr)
	}
	cx.downgrade(ctx, "k", testErr)
}

func TestCacheX_mDowngrade(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test")
	k := []string{"k_1", "k_2", "k_3"}
	cx := CacheX[string, string]{
		logger: logger.NewDefaultLogger(),
	}
	cx.mDowngrade(ctx, k, testErr)
	cx.mDowngradeCallback = func(ctx context.Context, keys []string, err error) {
		assert.EqualValues(t, k, keys)
		assert.ErrorIs(t, err, testErr)
	}
}
