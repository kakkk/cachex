package cachex

import "errors"

// ErrNotFound 回源查不到数据返回错误
var ErrNotFound = errors.New("not found")

type CacheError interface {
	Error() string
	GetErrorByLevel(level int) error
	GetErrorLevels() map[int]bool
}
