package cachex

import (
	"context"
	"fmt"
	"time"

	"github.com/kakkk/cachex/cache"
	"github.com/kakkk/cachex/internal/logger"
)

type Builder[K comparable, V any] struct {
	ctx context.Context
	cx  *CacheX[K, V]
}

// NewBuilder NewBuilder
func NewBuilder[K comparable, V any](ctx context.Context) *Builder[K, V] {
	return &Builder[K, V]{
		ctx: ctx,
		cx:  &CacheX[K, V]{},
	}
}

// SetName 设置缓存名称
func (b *Builder[K, V]) SetName(name string) *Builder[K, V] {
	b.cx.name = name
	return b
}

// AddCache 添加多级缓存
func (b *Builder[K, V]) AddCache(cache cache.Cache[V]) *Builder[K, V] {
	b.cx.caches = append(b.cx.caches, cache)
	return b
}

// SetGetDataKey 设置获取DataKey函数
func (b *Builder[K, V]) SetGetDataKey(fn GetDataKey[K]) *Builder[K, V] {
	b.cx.getDataKey = fn
	return b
}

// SetGetRealData 设置回源函数
func (b *Builder[K, V]) SetGetRealData(fn GetRealData[K, V]) *Builder[K, V] {
	b.cx.getRealData = fn
	return b
}

// SetMGetRealData 设置批量回源函数
func (b *Builder[K, V]) SetMGetRealData(fn MGetRealData[K, V]) *Builder[K, V] {
	b.cx.mGetRealData = fn
	return b
}

// SetHitCallback 设置缓存命中回调
func (b *Builder[K, V]) SetHitCallback(fn HitCallback) *Builder[K, V] {
	b.cx.hitCallback = fn
	return b
}

// SetMHitCallback 设置缓存批量命中回调
func (b *Builder[K, V]) SetMHitCallback(fn MHitCallback) *Builder[K, V] {
	b.cx.mHitCallback = fn
	return b
}

// SetLogger 设置Logger
func (b *Builder[K, V]) SetLogger(logger Logger) *Builder[K, V] {
	b.cx.logger = logger
	return b
}

// SetAllowDowngrade 设置是否允许降级
func (b *Builder[K, V]) SetAllowDowngrade(allow bool) *Builder[K, V] {
	b.cx.allowDowngrade = allow
	return b
}

// SetDowngradeCacheExpireTime 设置降级最大业务过期时间
func (b *Builder[K, V]) SetDowngradeCacheExpireTime(t time.Duration) *Builder[K, V] {
	b.cx.downgradeCacheExpireTime = t
	return b
}

// SetDowngradeCallBack 设置单个降级回调降级
func (b *Builder[K, V]) SetDowngradeCallBack(cb DowngradeCallBack[K]) *Builder[K, V] {
	b.cx.downgradeCallback = cb
	return b
}

// SetMDowngradeCallBack 设置批量降级回调降级
func (b *Builder[K, V]) SetMDowngradeCallBack(cb MDowngradeCallBack[K]) *Builder[K, V] {
	b.cx.mDowngradeCallback = cb
	return b
}

func (b *Builder[K, V]) SetIsSetDefault(isSetDefault bool) *Builder[K, V] {
	b.cx.isSetDefault = isSetDefault
	return b
}

// Build 设置并初始化缓存
func (b *Builder[K, V]) Build() (*CacheX[K, V], error) {
	// 设置logger
	if b.cx.logger == nil {
		b.cx.logger = logger.NewDefaultLogger()
	}
	// 检查GetDataKey
	if b.cx.getDataKey == nil {
		b.cx.logger.Errorf(b.ctx, "GetDataKey not set")
		return nil, fmt.Errorf("GetDataKey not set")
	}
	// cache可用性检测
	for level := 0; level < len(b.cx.caches); level++ {
		pong, err := b.cx.caches[level].Ping(b.ctx)
		if err != nil {
			b.cx.logger.Errorf(b.ctx, "cache %v level %v cache access fail: [%v]", b.cx.name, level, err)
			return nil, fmt.Errorf("cache access fail: [%w]", err)
		}
		b.cx.logger.Debugf(b.ctx, "cache %v level %v cache access, ping: %v", b.cx.name, level, pong)
	}
	// 初始化成功
	b.cx.logger.Debugf(b.ctx, "cache %v check success", b.cx.name)
	return b.cx, nil
}
