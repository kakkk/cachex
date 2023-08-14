package errors

import (
	"strconv"
	"strings"

	"github.com/kakkk/cachex/internal/utils"
)

type CacheErrorImpl struct {
	errors map[int]error
}

func NewCacheSetError() *CacheErrorImpl {
	var err *CacheErrorImpl
	return err
}

func (e *CacheErrorImpl) Error() string {
	if e == nil || len(e.errors) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("CacheError:[")
	for level, err := range e.errors {
		builder.WriteString("[level:")
		builder.WriteString(strconv.Itoa(level))
		builder.WriteString(",error:")
		builder.WriteString(err.Error())
		builder.WriteString("]")
	}
	builder.WriteString("]")
	return builder.String()
}

func (e *CacheErrorImpl) GetErrorByLevel(level int) error {
	if e == nil || e.errors == nil {
		return nil
	}
	err, ok := e.errors[level]
	if !ok || err == nil {
		return nil
	}
	return err
}

func (e *CacheErrorImpl) GetErrorLevels() map[int]bool {
	if e == nil {
		return make(map[int]bool)
	}
	return utils.GetMapKeys(e.errors)
}

func (e *CacheErrorImpl) AppendError(level int, err error) *CacheErrorImpl {
	if e == nil {
		return &CacheErrorImpl{
			errors: map[int]error{level: err},
		}
	}
	e.errors[level] = err
	return e
}
