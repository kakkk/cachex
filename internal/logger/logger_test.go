package logger

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogger(t *testing.T) {
	ctx := context.Background()
	l := NewDefaultLogger()
	buf := bytes.Buffer{}
	log.SetOutput(&buf)
	l.Debugf(ctx, "test")
	assert.Contains(t, buf.String(), "[DEBUG] test\n")
	buf.Reset()
	l.Infof(ctx, "test")
	assert.Contains(t, buf.String(), "[INFO] test\n")
	buf.Reset()
	l.Warnf(ctx, "test")
	assert.Contains(t, buf.String(), "[WARN] test\n")
	buf.Reset()
	l.Errorf(ctx, "test")
	assert.Contains(t, buf.String(), "[ERROR] test\n")

}
