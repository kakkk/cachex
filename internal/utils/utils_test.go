package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kakkk/cachex/internal/model"
)

func TestDuplicate(t *testing.T) {
	type args[T comparable] struct {
		list []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			args: args[int]{[]int{1, 2, 2, 3, 4, 5, 5, 6}},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			args: args[int]{[]int{1, 2, 3, 4, 5, 6, 7}},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			args: args[int]{[]int{1, 1, 1, 1, 1, 1, 1, 1}},
			want: []int{1},
		},
		{
			args: args[int]{[]int{}},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Duplicate(tt.args.list), "Duplicate(%v)", tt.args.list)
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	type args[K comparable, V any] struct {
		m map[K]V
	}
	type testCase[K comparable, V any] struct {
		name string
		args args[K, V]
		want map[K]bool
	}
	tests := []testCase[int, int]{
		{
			args: args[int, int]{
				m: map[int]int{
					1: 1,
					2: 0,
					3: 0,
				},
			},
			want: map[int]bool{
				1: true,
				2: true,
				3: true,
			},
		},
		{
			args: args[int, int]{
				m: map[int]int{},
			},
			want: map[int]bool{},
		},
		{
			args: args[int, int]{
				m: nil,
			},
			want: map[int]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetMapKeys(tt.args.m), "GetMapKeys(%v)", tt.args.m)
		})
	}
}

func TestMergeData(t *testing.T) {
	type args[K comparable, V any] struct {
		data map[K]V
		m    map[K]V
	}
	type testCase[K comparable, V any] struct {
		name string
		args args[K, V]
		want map[K]V
	}
	tests := []testCase[string, string]{
		{
			args: args[string, string]{
				data: map[string]string{
					"k_1": "v_1",
					"k_2": "v_2",
				},
				m: map[string]string{
					"k_3": "v_3",
				},
			},
			want: map[string]string{
				"k_1": "v_1",
				"k_2": "v_2",
				"k_3": "v_3",
			},
		},
		{
			args: args[string, string]{
				data: map[string]string{},
				m:    map[string]string{"k": "v"},
			},
			want: map[string]string{"k": "v"},
		},
		{
			args: args[string, string]{
				data: nil,
				m:    map[string]string{"k": "v"},
			},
			want: map[string]string{"k": "v"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MergeData(tt.args.data, tt.args.m), "MergeData(%v, %v)", tt.args.data, tt.args.m)
		})
	}
}

func TestConvertCacheDataMap(t *testing.T) {
	type args[K comparable, V any] struct {
		keys   []K
		data   map[string]V
		getKey func(key K) string
	}
	type testCase[K comparable, V any] struct {
		name string
		args args[K, V]
		want map[K]V
	}
	tests := []testCase[string, string]{
		{
			args: args[string, string]{
				keys: []string{"k_1", "k_2", "k_3", "k_4"},
				data: map[string]string{
					"k_k_1": "v_1",
					"k_k_2": "v_2",
					"k_k_3": "v_3",
				},
				getKey: func(key string) string {
					return fmt.Sprintf("k_%s", key)
				},
			},
			want: map[string]string{
				"k_1": "v_1",
				"k_2": "v_2",
				"k_3": "v_3",
			},
		},
		{
			args: args[string, string]{
				keys:   nil,
				data:   nil,
				getKey: nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ConvertCacheDataMap(tt.args.keys, tt.args.data, tt.args.getKey), "ConvertCacheDataMap(%v, %v, %v)", tt.args.keys, tt.args.data, tt.args.getKey)
		})
	}
}

func TestUnmarshalData(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
		str := `{"c":1017072000000,"d":"test"}`
		want := &model.CacheData[string]{
			CreateAt: 1017072000000,
			Data:     "test",
		}
		got, err := UnmarshalData[string]([]byte(str))
		assert.Nil(tt, err)
		assert.EqualValues(tt, want, got)
	})
	t.Run("unmarshal error", func(tt *testing.T) {
		str := `{`
		got, err := UnmarshalData[string]([]byte(str))
		assert.NotNil(tt, err)
		assert.Nil(tt, got)
	})
}

func TestMarshalData(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
		got, err := MarshalData("test", 1017072000000)
		assert.Nil(tt, err)
		assert.JSONEq(tt, `{"c":1017072000000,"d":"test"}`, string(got))
	})
	t.Run("marshal error", func(tt *testing.T) {
		var ch chan string
		got, err := MarshalData(ch, 1017072000000)
		assert.NotNil(tt, err)
		assert.Nil(tt, got)
	})
}

func TestNewData(t *testing.T) {
	got := NewData[string]("test", 1017072000000)
	assert.NotNil(t, got)
	assert.Equal(t, got.Data, "test")
	assert.Equal(t, got.CreateAt, int64(1017072000000))
}
