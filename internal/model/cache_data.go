package model

type CacheData[T any] struct {
	CreateAt int64 `json:"c"`
	Data     T     `json:"d"`
}
