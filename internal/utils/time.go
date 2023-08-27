package utils

import (
	"math/rand"
	"time"
)

func ConvertTimestamp(t time.Time) int64 {
	return t.UnixMilli()
}

// IsExpired 检查是否过期
func IsExpired(createAt int64, now time.Time, expire time.Duration) bool {
	// expire小于等于0，不过期
	if expire <= 0 {
		return false
	}
	// 创建时间+业务过期时间小于当前时间, 已过期
	if createAt+expire.Milliseconds() < now.UnixMilli() {
		return true
	}
	return false
}

func GetRandomTTL() time.Duration {
	return time.Duration(rand.Int()%200) * time.Millisecond
}
