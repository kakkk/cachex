package cachex

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/cache"
	"github.com/kakkk/cachex/internal/logger"
)

func TestBuilder(t *testing.T) {
	name := "test"
	cache0 := cache.NewCacheMocker[string]()
	cache1 := cache.NewCacheMocker[string]()
	getDataKey := func(_ string) string { return "test" }
	getRealData := func(_ context.Context, _ string) (string, error) { return "", nil }
	mGetRealData := func(_ context.Context, _ []string) (map[string]string, error) { return nil, nil }
	hitCallback := func(_ string, _ int) {}
	mHitCallback := func(_ string, _ int, _ int) {}
	log := logger.NewDefaultLogger()
	downgradeExpireTime := time.Hour
	downgradeCallBack := func(_ context.Context, _ string, _ error) {}
	mDowngradeCallBack := func(_ context.Context, _ []string, _ error) {}

	t.Run("set_all", func(tt *testing.T) {
		cx, err := NewBuilder[string, string](context.Background()).
			SetName(name).
			AddCache(cache0).
			AddCache(cache1).
			SetGetDataKey(getDataKey).
			SetGetRealData(getRealData).
			SetMGetRealData(mGetRealData).
			SetHitCallback(hitCallback).
			SetMHitCallback(mHitCallback).
			SetLogger(log).
			SetAllowDowngrade(true).
			SetDowngradeCacheExpireTime(downgradeExpireTime).
			SetDowngradeCallBack(downgradeCallBack).
			SetMDowngradeCallBack(mDowngradeCallBack).
			SetIsSetDefault(true).
			Build()

		assert.Nil(t, err)
		assert.Equal(tt, name, cx.name)
		assert.Equal(tt, 2, len(cx.caches))
		assert.NotNil(tt, cx.getDataKey)
		assert.NotNil(tt, cx.getRealData)
		assert.NotNil(tt, cx.mGetRealData)
		assert.NotNil(tt, cx.hitCallback)
		assert.NotNil(tt, cx.mHitCallback)
		assert.NotNil(tt, cx.logger)
		assert.True(tt, cx.allowDowngrade)
		assert.Equal(tt, downgradeExpireTime, cx.downgradeCacheExpireTime)
		assert.NotNil(tt, cx.downgradeCallback)
		assert.NotNil(tt, cx.mDowngradeCallback)
		assert.True(tt, cx.isSetDefault)
	})

	t.Run("not_set_logger", func(tt *testing.T) {
		cx, err := NewBuilder[string, string](context.Background()).
			SetName(name).
			AddCache(cache0).
			SetGetDataKey(getDataKey).
			Build()
		assert.Nil(t, err)
		assert.NotNil(tt, cx.logger)
	})

	t.Run("get_data_key_not_set", func(tt *testing.T) {
		_, err := NewBuilder[string, string](context.Background()).
			SetName(name).
			AddCache(cache0).
			Build()
		assert.NotNil(tt, err)
	})

	t.Run("cache_access_fail", func(tt *testing.T) {
		cacheNotAccess := cache.NewCacheMocker[string]().
			MockPing(func(_ context.Context) (string, error) {
				return "", errors.New("unit_test")
			})
		_, err := NewBuilder[string, string](context.Background()).
			SetName(name).
			AddCache(cacheNotAccess).
			SetGetDataKey(getDataKey).
			Build()
		assert.NotNil(tt, err)
	})
}
