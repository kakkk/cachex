package cache

import (
	"context"
	"time"
)

type Mocker[T any] struct {
	mockGet     func(ctx context.Context, key string, expire time.Duration) (T, bool)
	mockMGet    func(ctx context.Context, keys []string, expire time.Duration) map[string]T
	mockSet     func(ctx context.Context, key string, data T, createTime time.Time) error
	mockMSet    func(ctx context.Context, kvs map[string]T, createTime time.Time) error
	mockDelete  func(ctx context.Context, key string) error
	mockMDelete func(ctx context.Context, keys []string) error
	mockPing    func(ctx context.Context) (string, error)
}

func NewCacheMocker[T any]() *Mocker[T] {
	return &Mocker[T]{}
}

func (m *Mocker[T]) Get(ctx context.Context, key string, expire time.Duration) (T, bool) {
	if m.mockGet != nil {
		return m.mockGet(ctx, key, expire)
	}
	var zero T
	return zero, false
}

func (m *Mocker[T]) MGet(ctx context.Context, keys []string, expire time.Duration) map[string]T {
	if m.mockMGet != nil {
		return m.mockMGet(ctx, keys, expire)
	}
	return make(map[string]T)
}

func (m *Mocker[T]) Set(ctx context.Context, key string, data T, createTime time.Time) error {
	if m.mockSet != nil {
		return m.mockSet(ctx, key, data, createTime)
	}
	return nil
}

func (m *Mocker[T]) MSet(ctx context.Context, kvs map[string]T, createTime time.Time) error {
	if m.mockMSet != nil {
		return m.mockMSet(ctx, kvs, createTime)
	}
	return nil
}

func (m *Mocker[T]) Delete(ctx context.Context, key string) error {
	if m.mockDelete != nil {
		return m.mockDelete(ctx, key)
	}
	return nil
}

func (m *Mocker[T]) MDelete(ctx context.Context, keys []string) error {
	if m.mockMDelete != nil {
		return m.mockMDelete(ctx, keys)
	}
	return nil
}

func (m *Mocker[T]) Ping(ctx context.Context) (string, error) {
	if m.mockPing != nil {
		return m.mockPing(ctx)
	}
	return "Pong", nil
}

func (m *Mocker[T]) MockGet(mockFn func(ctx context.Context, key string, expire time.Duration) (T, bool)) *Mocker[T] {
	m.mockGet = mockFn
	return m
}

func (m *Mocker[T]) MockMGet(mockFn func(ctx context.Context, keys []string, expire time.Duration) map[string]T) *Mocker[T] {
	m.mockMGet = mockFn
	return m
}

func (m *Mocker[T]) MockSet(mockFn func(ctx context.Context, key string, data T, createTime time.Time) error) *Mocker[T] {
	m.mockSet = mockFn
	return m
}

func (m *Mocker[T]) MockMSet(mockFn func(ctx context.Context, kvs map[string]T, createTime time.Time) error) *Mocker[T] {
	m.mockMSet = mockFn
	return m
}

func (m *Mocker[T]) MockDelete(mockFn func(ctx context.Context, key string) error) *Mocker[T] {
	m.mockDelete = mockFn
	return m
}

func (m *Mocker[T]) MockMDelete(mockFn func(ctx context.Context, keys []string) error) *Mocker[T] {
	m.mockMDelete = mockFn
	return m
}

func (m *Mocker[T]) MockPing(mockFn func(ctx context.Context) (string, error)) *Mocker[T] {
	m.mockPing = mockFn
	return m
}
