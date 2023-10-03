package model

type CacheData[T any] struct {
	CreateAt int64 `json:"c"`
	Data     T     `json:"d"`
	Default  uint  `json:"z"`
}

func (c *CacheData[T]) IsDefault() bool {
	return c.Default == 1
}
