package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertTimestamp(t *testing.T) {
	now := time.Now()
	got := ConvertTimestamp(now)
	assert.Equal(t, now.UnixMilli(), got)
}

func TestIsExpired(t *testing.T) {
	type args struct {
		createAt int64
		now      time.Time
		expire   time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "never expired 0",
			args: args{
				createAt: time.Now().UnixMilli(),
				now:      time.Now(),
				expire:   0,
			},
			want: false,
		},
		{
			name: "never expired 1",
			args: args{
				createAt: time.Now().UnixMilli(),
				now:      time.Now().Add(-time.Hour),
				expire:   -1,
			},
			want: false,
		},
		{
			name: "not expired",
			args: args{
				createAt: time.Now().Add(-time.Hour).UnixMilli(),
				now:      time.Now(),
				expire:   2 * time.Hour,
			},
			want: false,
		},
		{
			name: "expired",
			args: args{
				createAt: time.Now().Add(-2 * time.Hour).UnixMilli(),
				now:      time.Now(),
				expire:   time.Hour,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsExpired(tt.args.createAt, tt.args.now, tt.args.expire), "IsExpired(%v, %v, %v)", tt.args.createAt, tt.args.now, tt.args.expire)
		})
	}
}
