package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheSetError(t *testing.T) {
	err := NewCacheSetError()
	assert.Nil(t, err)
}

func TestCacheErrorImpl_Error(t *testing.T) {
	t.Run("normal", func(tt *testing.T) {
		e := &CacheErrorImpl{
			errors: map[int]error{
				0: errors.New("l0_err"),
				1: errors.New("l1_err"),
			},
		}
		got := e.Error()
		assert.Contains(tt, got, "CacheError:")
		assert.Contains(tt, got, "[level:0,error:l0_err]")
		assert.Contains(tt, got, "[level:1,error:l1_err]")
	})

	t.Run("nil or empty", func(tt *testing.T) {
		e := &CacheErrorImpl{
			errors: map[int]error{},
		}
		got := e.Error()
		assert.Equal(tt, "", got)

		var e2 *CacheErrorImpl
		got = e2.Error()
		assert.Equal(tt, "", got)
	})
}

func TestCacheErrorImpl_GetErrorByLevel(t *testing.T) {
	t.Run("nil or empty", func(tt *testing.T) {
		e := &CacheErrorImpl{
			errors: map[int]error{},
		}
		got := e.GetErrorByLevel(0)
		assert.Nil(tt, got)

		var e2 *CacheErrorImpl
		got = e2.GetErrorByLevel(0)
		assert.Nil(tt, got)
	})
	t.Run("normal", func(tt *testing.T) {
		l0Err := errors.New("l0_err")
		l1Err := errors.New("l1_err")
		e := &CacheErrorImpl{
			errors: map[int]error{
				0: l0Err,
				1: l1Err,
			},
		}
		got := e.GetErrorByLevel(0)
		assert.ErrorIs(tt, got, l0Err)
		got = e.GetErrorByLevel(1)
		assert.ErrorIs(tt, got, l1Err)
		got = e.GetErrorByLevel(2)
		assert.Nil(tt, got)
	})

}

func TestCacheErrorImpl_GetErrorLevels(t *testing.T) {
	var e *CacheErrorImpl
	got := e.GetErrorLevels()
	assert.EqualValues(t, map[int]bool{}, got)
	e = &CacheErrorImpl{
		errors: map[int]error{
			0: errors.New("l0_err"),
			1: errors.New("l1_err"),
		},
	}
	got = e.GetErrorLevels()
	want := map[int]bool{
		0: true,
		1: true,
	}
	assert.EqualValues(t, want, got)
}

func TestCacheErrorImpl_AppendError(t *testing.T) {
	var e *CacheErrorImpl
	l0Err := errors.New("l0_err")
	l1Err := errors.New("l1_err")
	got := e.AppendError(0, l0Err)
	assert.NotNil(t, got)
	assert.ErrorIs(t, got.errors[0], l0Err)
	got = got.AppendError(1, l1Err)
	assert.ErrorIs(t, got.errors[1], l1Err)
}
