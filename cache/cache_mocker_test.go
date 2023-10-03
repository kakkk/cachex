package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheMocker(t *testing.T) {
	ctx := context.Background()
	t.Run("not set", func(tt *testing.T) {
		mocker := NewCacheMocker[string]()
		got, ok := mocker.Get(ctx, "k", 0)
		assert.False(tt, ok)
		assert.Equal(tt, "", got)
		mGot := mocker.MGet(ctx, []string{"k_1", "k_2"}, 0)
		assert.Empty(tt, mGot)
		err := mocker.Set(ctx, "k", "v", time.Now())
		assert.Nil(tt, err)
		err = mocker.MSet(ctx, map[string]string{"k_1": "v_1", "k_2": "v_3"}, time.Now())
		assert.Nil(tt, err)
		err = mocker.SetDefault(ctx, []string{"k_1", "k_2"}, time.Now())
		assert.Nil(tt, err)
		err = mocker.Delete(ctx, "k")
		assert.Nil(tt, err)
		err = mocker.MDelete(ctx, []string{"k_1", "k_2"})
		assert.Nil(tt, err)
		pong, err := mocker.Ping(ctx)
		assert.Nil(tt, err)
		assert.Equal(tt, "Pong", pong)
	})
	t.Run("set all", func(tt *testing.T) {
		testErr := errors.New("test")
		mocker := NewCacheMocker[string]().
			MockGet(func(ctx context.Context, key string, expire time.Duration) (string, bool) {
				return "v", true
			}).
			MockMGet(func(ctx context.Context, keys []string, expire time.Duration) map[string]string {
				return map[string]string{"k_1": "v_1", "k_2": "v_3"}
			}).
			MockSet(func(ctx context.Context, key string, data string, createTime time.Time) error {
				return testErr
			}).
			MockMSet(func(ctx context.Context, kvs map[string]string, createTime time.Time) error {
				return testErr
			}).
			MockSetDefault(func(ctx context.Context, keys []string, createTime time.Time) error {
				return testErr
			}).
			MockDelete(func(ctx context.Context, key string) error {
				return testErr
			}).
			MockMDelete(func(ctx context.Context, keys []string) error {
				return testErr
			}).
			MockPing(func(ctx context.Context) (string, error) {
				return "", testErr
			})
		got, ok := mocker.Get(ctx, "k", 0)
		assert.True(tt, ok)
		assert.Equal(tt, "v", got)
		mGot := mocker.MGet(ctx, []string{"k_1", "k_2"}, 0)
		assert.EqualValues(tt, map[string]string{"k_1": "v_1", "k_2": "v_3"}, mGot)
		err := mocker.Set(ctx, "k", "v", time.Now())
		assert.ErrorIs(tt, err, testErr)
		err = mocker.MSet(ctx, map[string]string{"k_1": "v_1", "k_2": "v_3"}, time.Now())
		assert.ErrorIs(tt, err, testErr)
		err = mocker.SetDefault(ctx, []string{"k_1", "k_2"}, time.Now())
		assert.ErrorIs(tt, err, testErr)
		err = mocker.Delete(ctx, "k")
		assert.ErrorIs(tt, err, testErr)
		err = mocker.MDelete(ctx, []string{"k_1", "k_2"})
		assert.ErrorIs(tt, err, testErr)
		pong, err := mocker.Ping(ctx)
		assert.ErrorIs(tt, err, testErr)
		assert.Equal(tt, "", pong)
	})
}
