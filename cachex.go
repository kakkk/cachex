package cachex

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/kakkk/cachex/cache"
	"github.com/kakkk/cachex/internal/consts"
	cachexError "github.com/kakkk/cachex/internal/errors"
	"github.com/kakkk/cachex/internal/utils"
)

// GetDataKey 获取数据Key函数
type GetDataKey[K comparable] func(key K) string

// GetRealData 回源函数
type GetRealData[K comparable, V any] func(ctx context.Context, key K) (data V, err error)

// MGetRealData 批量回源函数
type MGetRealData[K comparable, V any] func(ctx context.Context, keys []K) (data map[K]V, err error)

// HitCallback 命中缓存回调函数
type HitCallback func(name string, level int)

// MHitCallback 批量命中缓存回调函数
type MHitCallback func(name string, level int, times int)

// DowngradeCallBack 降级回调函数
type DowngradeCallBack[K comparable] func(ctx context.Context, key K, err error)

// MDowngradeCallBack 批量降级回调函数
type MDowngradeCallBack[K comparable] func(ctx context.Context, keys []K, err error)

// CacheX CacheX组件
type CacheX[K comparable, V any] struct {
	name                     string                // 缓存名称
	caches                   []cache.Cache[V]      // 多级缓存
	getDataKey               GetDataKey[K]         // 获取缓存Key函数
	getRealData              GetRealData[K, V]     // 回源函数
	mGetRealData             MGetRealData[K, V]    // 批量回源函数
	hitCallback              HitCallback           // 命中回源
	mHitCallback             MHitCallback          // 批量命中回源
	logger                   Logger                // 自定义日志
	allowDowngrade           bool                  // 回源失败降级
	downgradeCacheExpireTime time.Duration         // 降级最大业务过期时间
	downgradeCallback        DowngradeCallBack[K]  // 降级回调
	mDowngradeCallback       MDowngradeCallBack[K] // 批量降级回调
}

// Set 设置缓存
func (cx *CacheX[K, V]) Set(ctx context.Context, key K, data V) (err error) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
			return
		}
	})()

	dataKey, now := cx.getDataKey(key), time.Now()
	setErrors := cachexError.NewCacheSetError()
	for level := 0; level < len(cx.caches); level++ {
		err := cx.caches[level].Set(ctx, dataKey, data, now)
		if err != nil {
			setErrors = setErrors.AppendError(level, err)
		}
	}
	return setErrors
}

// MSet 批量设置缓存
func (cx *CacheX[K, V]) MSet(ctx context.Context, kvs map[K]V) (err error) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
			return
		}
	})()
	if kvs == nil || len(kvs) == 0 {
		return nil
	}
	setErrors := cachexError.NewCacheSetError()
	now := time.Now()
	data := make(map[string]V, len(kvs))
	for k, v := range kvs {
		data[cx.getDataKey(k)] = v
	}
	for level := 0; level < len(cx.caches); level++ {
		err := cx.caches[level].MSet(ctx, data, now)
		if err != nil {
			setErrors = setErrors.AppendError(level, err)
		}
	}
	return setErrors
}

// Get 查询缓存
func (cx *CacheX[K, V]) Get(ctx context.Context, key K, expire time.Duration) (data V, ok bool) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			var zero V
			data, ok = zero, false
			return
		}
	})()
	dataKey := cx.getDataKey(key)
	// 查询缓存
	for level := len(cx.caches) - 1; level >= 0; level-- {
		var hit bool
		data, hit = cx.caches[level].Get(ctx, dataKey, expire)
		if hit {
			// 命中缓存，直接返回
			cx.hit(ctx, level)
			return data, true
		}
	}
	// 缓存失效，回源
	cx.hit(ctx, consts.CacheLevelSource)
	return cx.getRealDataInternal(ctx, key)
}

// MGet 批量查询缓存
func (cx *CacheX[K, V]) MGet(ctx context.Context, keys []K, expire time.Duration) (data map[K]V) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			data = make(map[K]V)
			return
		}
	})()
	data = make(map[K]V)
	// key去重
	keys = utils.Duplicate(keys)
	dataKeys := cx.mGetDataKeys(keys)

	// 从多级缓存中获取
	for level := len(cx.caches) - 1; level >= 0; level-- {
		got := cx.caches[level].MGet(ctx, dataKeys, expire)
		if len(got) != 0 {
			cx.mHit(ctx, level, len(got))
		}
		data = utils.MergeData(data, utils.ConvertCacheDataMap[K, V](keys, got, cx.getDataKey))
		// 当前data数量等于keys的数量，说明全部缓存已经命中，直接返回
		if len(data) == len(keys) {
			return data
		}
	}

	// 需要回源的keys
	var needGetRealDataKeys []K
	for _, key := range keys {
		if _, ok := data[key]; !ok {
			needGetRealDataKeys = append(needGetRealDataKeys, key)
		}
	}

	// 回源
	cx.mHit(ctx, consts.CacheLevelSource, len(needGetRealDataKeys))
	realData := cx.mGetRealDataInternal(ctx, needGetRealDataKeys)
	for k, v := range realData {
		data[k] = v
	}
	return data
}

// Delete 删除缓存
func (cx *CacheX[K, V]) Delete(ctx context.Context, key K) (err error) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
			return
		}
	})()
	dataKey := cx.getDataKey(key)
	delErrors := cachexError.NewCacheSetError()
	for level := 0; level < len(cx.caches); level++ {
		err := cx.caches[level].Delete(ctx, dataKey)
		if err != nil {
			delErrors = delErrors.AppendError(level, err)
		}
	}
	return delErrors
}

// Delete 批量删除缓存
func (cx *CacheX[K, V]) MDelete(ctx context.Context, keys []K) (err error) {
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
			return
		}
	})()
	dataKeys := cx.mGetDataKeys(keys)
	delErrors := cachexError.NewCacheSetError()
	for level := 0; level < len(cx.caches); level++ {
		err := cx.caches[level].MDelete(ctx, dataKeys)
		if err != nil {
			delErrors = delErrors.AppendError(level, err)
		}
	}
	return delErrors
}

// getRealDataInternal 回源
func (cx *CacheX[K, V]) getRealDataInternal(ctx context.Context, key K) (data V, ok bool) {
	var (
		err  error
		zero V
	)
	// 回源失败降级策略
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			data = zero
			ok = false
			// 不允许降级
			if !cx.allowDowngrade {
				return
			}
			// 降级查询缓存
			for level := len(cx.caches) - 1; level >= 0; level-- {
				data, ok = cx.caches[level].Get(ctx, cx.getDataKey(key), cx.downgradeCacheExpireTime)
				if ok {
					break
				}
			}
			cx.downgrade(ctx, key, err)
			return
		}
	})()

	// 没有配置回源，直接返回
	if cx.getRealData == nil {
		return zero, false
	}

	// 回源查询
	data, err = cx.getRealData(ctx, key)
	if err != nil {
		return
	}

	// 写入缓存
	_ = cx.Set(ctx, key, data)
	return data, true
}

// mGetRealDataInternal 批量回源
func (cx *CacheX[K, V]) mGetRealDataInternal(ctx context.Context, keys []K) (data map[K]V) {
	var err error
	defer cx.recover(ctx, func(r any) {
		if r != nil {
			err = fmt.Errorf("[panic recover] %v", r)
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			data = make(map[K]V)
			// 不允许降级
			if !cx.allowDowngrade {
				return
			}
			// 降级查询缓存, 从每一级获取缓存并组装
			dataKeys := cx.mGetDataKeys(keys)
			for level := len(cx.caches) - 1; level >= 0; level-- {
				got := cx.caches[level].MGet(ctx, dataKeys, cx.downgradeCacheExpireTime)
				data = utils.MergeData(data, utils.ConvertCacheDataMap[K, V](keys, got, cx.getDataKey))
				if len(data) == len(keys) {
					break
				}
			}
			// 回调
			cx.mDowngrade(ctx, keys, err)
			return
		}
	})()

	// 没有配置回源，直接返回
	if cx.mGetRealData == nil {
		return make(map[K]V)
	}

	// 回源查询
	data, err = cx.mGetRealData(ctx, keys)
	if err != nil {
		return
	}

	// 写入缓存
	_ = cx.MSet(ctx, data)
	return data

}

// mGetDataKeys 批量获取DataKey
func (cx *CacheX[K, V]) mGetDataKeys(keys []K) []string {
	dataKeys := make([]string, len(keys))
	for i := range keys {
		dataKeys[i] = cx.getDataKey(keys[i])
	}
	return dataKeys
}

// hit 命中回调
func (cx *CacheX[K, V]) hit(ctx context.Context, level int) {
	defer cx.recover(ctx, nil)()
	if cx.hitCallback != nil {
		cx.hitCallback(cx.name, level)
		return
	}
	cx.logger.Debugf(ctx, "cache %v hit, level:%v", cx.name, level)
	return
}

// mHit 批量命中回调
func (cx *CacheX[K, V]) mHit(ctx context.Context, level int, times int) {
	defer cx.recover(ctx, nil)()
	if cx.mHitCallback != nil {
		cx.mHitCallback(cx.name, level, times)
		return
	}
	cx.logger.Debugf(ctx, "cache %v hit, level:%v, times:%v", cx.name, level, times)
	return
}

// downgrade 降级回调
func (cx *CacheX[K, V]) downgrade(ctx context.Context, key K, err error) {
	defer cx.recover(ctx, nil)()
	if cx.downgradeCallback != nil {
		cx.downgradeCallback(ctx, key, err)
		return
	}
	cx.logger.Warnf(ctx, "cache downgrade, key:%v, error:%v", key, err)
}

// mDowngrade 批量降级回调
func (cx *CacheX[K, V]) mDowngrade(ctx context.Context, keys []K, err error) {
	defer cx.recover(ctx, nil)()
	if cx.mDowngradeCallback != nil {
		cx.mDowngradeCallback(ctx, keys, err)
		return
	}
	cx.logger.Warnf(ctx, "cache downgrade, keys:%v, error:%v", keys, err)
}

// recover Panic Recover
func (cx *CacheX[K, V]) recover(ctx context.Context, fn func(r any)) func() {
	return func() {
		r := recover()
		if r != nil {
			cx.logger.Errorf(ctx, "[panic recover] %v\nstack:\n%v", r, string(debug.Stack()))
		}
		if fn != nil {
			fn(r)
		}
		return
	}
}
