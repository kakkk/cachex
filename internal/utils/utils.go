package utils

import (
	"github.com/kakkk/cachex/internal/json"
	"github.com/kakkk/cachex/internal/model"
)

func Duplicate[T comparable](list []T) []T {
	set := make(map[T]struct{})
	if len(list) == 0 {
		return list
	}
	res := make([]T, 0, len(list))
	for i := 0; i < len(list); i++ {
		if _, ok := set[list[i]]; ok {
			continue
		}
		res = append(res, list[i])
		set[list[i]] = struct{}{}
	}
	return res
}

func GetMapKeys[K comparable, V any](m map[K]V) map[K]bool {
	keys := make(map[K]bool, len(m))
	for k := range m {
		keys[k] = true
	}
	return keys
}

func MergeData[K comparable, V any](data map[K]V, m map[K]V) map[K]V {
	if len(data) == 0 {
		return m
	}
	for k, v := range m {
		if _, ok := data[k]; !ok {
			data[k] = v
		}
	}
	return data
}

func ConvertCacheDataMap[K comparable, V any](keys []K, data map[string]V, getKey func(key K) string) map[K]V {
	res := make(map[K]V, len(data))
	for _, key := range keys {
		if v, ok := data[getKey(key)]; ok {
			res[key] = v
		}
	}
	return res
}

func UnmarshalData[T any](val []byte) (*model.CacheData[T], error) {
	var zero *model.CacheData[T]
	data := &model.CacheData[T]{}
	err := json.Unmarshal(val, data)
	if err != nil {
		return zero, err
	}
	return data, err
}

func MarshalData[T any](data T, createAt int64) ([]byte, error) {
	cacheData := &model.CacheData[T]{
		CreateAt: createAt,
		Data:     data,
	}
	return json.Marshal(cacheData)
}
